package sequencer

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/sig-0/go-ibft"
	"github.com/sig-0/go-ibft/message/types"
)

// Sequencer is the consensus actor's (ibft.Validator) block finalization process. Whenever the network moves to a
// new sequence, all actors run their Sequencer processes to reach consensus on some proposal. Sequences consist of
// rounds in which a chosen actor (Proposer) suggests their own proposal to the network. The Sequencer makes sure
// that consensus is (eventually) reached, moving to higher rounds in case the network cannot agree on some proposal.
// Given its simple API method Finalize, Sequencer is designed to work alongside a syncing protocol
type Sequencer struct {
	validator      ibft.Validator
	validatorSet   ibft.ValidatorSet
	state          state
	wg             sync.WaitGroup
	round0Duration time.Duration // Maximum time for the initial round of consensus. MUST be network-wide!
}

// NewSequencer returns a Sequencer object for the provided validator
func NewSequencer(v ibft.Validator, vs ibft.ValidatorSet, round0Duration time.Duration) *Sequencer {
	return &Sequencer{
		validator:      v,
		validatorSet:   vs,
		round0Duration: round0Duration,
	}
}

// Finalize runs the block finalization loop. This method returns a non-nil value only if consensus
// is reached for the provided sequence. Otherwise, it runs forever until cancelled by the caller
func (s *Sequencer) Finalize(ctx Context, sequence uint64) *types.FinalizedProposal {
	s.state.init(sequence)

	c := make(chan *types.FinalizedProposal, 1)
	go func() {
		defer close(c)

		fb := s.finalize(ctx)
		if fb == nil {
			return
		}

		c <- fb
	}()

	select {
	case <-ctx.Done():
		<-c // wait for finalize to return

		return nil
	case fb := <-c:
		return fb
	}
}

// finalize starts the round runner loop. In each round (loop iteration), 4 processes run in parallel.
// This method returns only if the block finalization algorithm is complete or if the caller cancelled the Context
func (s *Sequencer) finalize(ctx Context) *types.FinalizedProposal {
	for {
		c, cancelRound := context.WithCancel(ctx)
		ctxRound := Context{c}
		teardown := func() {
			cancelRound()
			s.wg.Wait()
		}

		select {
		case _, ok := <-s.startRoundTimer(ctxRound):
			teardown()
			if !ok {
				return nil
			}

			s.state.moveToNextRound()
			s.sendMsgRoundChange(ctx)

		case rcc, ok := <-s.awaitHigherRoundRCC(ctxRound):
			teardown()
			if !ok {
				return nil
			}

			s.state.acceptRCC(rcc)

		case proposal, ok := <-s.awaitHigherRoundProposal(ctxRound):
			teardown()
			if !ok {
				return nil
			}

			s.state.acceptProposal(proposal)
			s.sendMsgPrepare(ctx)

		case fb, ok := <-s.awaitFinalizedBlockInCurrentRound(ctxRound):
			teardown()
			if !ok {
				return nil
			}

			return fb
		}
	}
}

// startRoundTimer starts the round timer of the current round
func (s *Sequencer) startRoundTimer(ctx Context) <-chan struct{} {
	s.wg.Add(1)

	c := make(chan struct{}, 1)
	go func(view *types.View) {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		roundTimer := s.getRoundTimer(view.Round)

		select {
		case <-ctx.Done():
			roundTimer.Stop()
		case <-roundTimer.C:
			c <- struct{}{}
		}
	}(s.state.getView())

	return c
}

// awaitHigherRoundProposal listens for proposal messages from rounds higher than the current
func (s *Sequencer) awaitHigherRoundProposal(ctx Context) <-chan *types.MsgProposal {
	s.wg.Add(1)

	c := make(chan *types.MsgProposal, 1)
	go func(view *types.View) {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		proposal, err := s.awaitProposal(ctx, view, true)
		if err != nil {
			return
		}

		c <- proposal
	}(s.state.getView())

	return c
}

// awaitHigherRoundRCC listens for round change certificates from rounds higher than the current
func (s *Sequencer) awaitHigherRoundRCC(ctx Context) <-chan *types.RoundChangeCertificate {
	s.wg.Add(1)

	c := make(chan *types.RoundChangeCertificate, 1)
	go func(view *types.View) {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		rcc, err := s.awaitRCC(ctx, view, true)
		if err != nil {
			return
		}

		c <- rcc
	}(s.state.getView())

	return c
}

// awaitFinalizedBlockInCurrentRound starts the block finalization algorithm for the current round
func (s *Sequencer) awaitFinalizedBlockInCurrentRound(ctx Context) <-chan *types.FinalizedProposal {
	s.wg.Add(1)

	c := make(chan *types.FinalizedProposal, 1)
	go func() {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		if s.shouldPropose() {
			p, err := s.buildProposal(ctx)
			if err != nil {
				return
			}

			s.sendMsgProposal(ctx, p)
		}

		if !s.state.isProposalAccepted() {
			if err := s.awaitCurrentRoundProposal(ctx); err != nil {
				return
			}

			s.sendMsgPrepare(ctx)
		}

		if err := s.awaitPrepare(ctx); err != nil {
			return
		}

		s.sendMsgCommit(ctx)

		if err := s.awaitCommit(ctx); err != nil {
			return
		}

		c <- s.state.getFinalizedProposal()
	}()

	return c
}

func (s *Sequencer) getRoundTimer(round uint64) *time.Timer {
	return time.NewTimer(s.round0Duration * time.Duration(math.Pow(2, float64(round))))
}

func (s *Sequencer) shouldPropose() bool {
	return s.validatorSet.IsProposer(s.validator.ID(), s.state.getSequence(), s.state.getRound())
}

func (s *Sequencer) buildProposal(ctx Context) ([]byte, error) {
	if s.state.getRound() == 0 {
		return s.validator.BuildProposal(s.state.getSequence()), nil
	}

	if s.state.rcc == nil {
		// round jump triggered by round timer -> justify proposal with rcc
		RCC, err := s.awaitRCC(ctx, s.state.getView(), false)
		if err != nil {
			return nil, err
		}

		s.state.rcc = RCC
	}

	block, _ := s.state.rcc.HighestRoundBlock()
	if block == nil {
		return s.validator.BuildProposal(s.state.getSequence()), nil
	}

	return block, nil
}
