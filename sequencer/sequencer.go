package sequencer

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

// MessageFeed provides an asynchronous way to receive consensus messages. In addition
// to listen for any type of message for any particular view, the higherRounds flag provides an option
// to receive messages from rounds higher than the round in provided view.
//
// CONTRACT:
//
// 1. any message is valid:
//   - no required fields missing (Sender, Signature, View)
//   - signature is valid [ recoverFrom(msg.Payload, Signature) == Sender ]
//
// 2. all messages are considered unique (there cannot be 2 or more messages with identical From fields)
type MessageFeed interface {
	// ProposalMessages returns the MsgProposal subscription for given view(s)
	ProposalMessages(view *types.View, higherRounds bool) (types.Subscription[*types.MsgProposal], func())

	// PrepareMessages returns the MsgPrepare subscription for given view(s)
	PrepareMessages(view *types.View, higherRounds bool) (types.Subscription[*types.MsgPrepare], func())

	// CommitMessages returns the MsgCommit subscription for given view(s)
	CommitMessages(view *types.View, higherRounds bool) (types.Subscription[*types.MsgCommit], func())

	// RoundChangeMessages returns the MsgRoundChange subscription for given view(s)
	RoundChangeMessages(view *types.View, higherRounds bool) (types.Subscription[*types.MsgRoundChange], func())
}

type MessageTransport struct {
	Proposal    ibft.Transport[*types.MsgProposal]
	Prepare     ibft.Transport[*types.MsgPrepare]
	Commit      ibft.Transport[*types.MsgCommit]
	RoundChange ibft.Transport[*types.MsgRoundChange]
}

// Sequencer is an IBFT-Context based block finalizer of a consensus actor. Its only purpose
// is to determine if consensus is reached in a particular sequence (height)
// and output the finalized block. Because of its simple API it's envisioned to
// work alongside a block syncing process
type Sequencer struct {
	ibft.Validator

	state viewState
	wg    sync.WaitGroup

	Round0Duration time.Duration
}

// New instantiates a new Sequencer object
func New(val ibft.Validator, round0Duration time.Duration) *Sequencer {
	return &Sequencer{
		Validator:      val,
		Round0Duration: round0Duration,
	}
}

// Finalize runs the block finalization loop. This method returns a non-nil value only if
// consensus is reached for the given sequence. Otherwise, it runs forever until cancelled by the caller
func (s *Sequencer) Finalize(ctx Context, sequence uint64) *types.FinalizedProposal {
	s.state.Init(sequence)

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

			s.state.MoveToNextRound()
			s.sendMsgRoundChange(ctx)

		case rcc, ok := <-s.awaitHigherRoundRCC(ctxRound):
			teardown()

			if !ok {
				return nil
			}

			s.state.AcceptRCC(rcc)

		case proposal, ok := <-s.awaitHigherRoundProposal(ctxRound):
			teardown()

			if !ok {
				return nil
			}

			s.state.AcceptProposal(proposal)
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
	c := make(chan struct{}, 1)

	s.wg.Add(1)

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
	}(s.state.View())

	return c
}

// awaitHigherRoundProposal listens for proposal messages from rounds higher than the current
func (s *Sequencer) awaitHigherRoundProposal(ctx Context) <-chan *types.MsgProposal {
	c := make(chan *types.MsgProposal, 1)

	s.wg.Add(1)

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
	}(s.state.View())

	return c
}

// awaitHigherRoundRCC listens for round change certificates from rounds higher than the current
func (s *Sequencer) awaitHigherRoundRCC(ctx Context) <-chan *types.RoundChangeCertificate {
	c := make(chan *types.RoundChangeCertificate, 1)

	s.wg.Add(1)

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
	}(s.state.View())

	return c
}

// awaitFinalizedBlockInCurrentRound starts the block finalization algorithm for the current round
func (s *Sequencer) awaitFinalizedBlockInCurrentRound(ctx Context) <-chan *types.FinalizedProposal {
	c := make(chan *types.FinalizedProposal, 1)

	s.wg.Add(1)

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

		if !s.state.ProposalAccepted() {
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

		c <- s.state.FinalizedProposal()
	}()

	return c
}

func (s *Sequencer) getRoundTimer(round uint64) *time.Timer {
	return time.NewTimer(s.Round0Duration * time.Duration(math.Pow(2, float64(round))))
}

func (s *Sequencer) shouldPropose() bool {
	return s.IsProposer(s.ID(), s.state.Sequence(), s.state.Round())
}

func (s *Sequencer) buildProposal(ctx Context) ([]byte, error) {
	if s.state.Round() == 0 {
		return s.BuildProposal(s.state.Sequence()), nil
	}

	if s.state.rcc == nil {
		// round jump triggered by round timer -> justify proposal with rcc
		RCC, err := s.awaitRCC(ctx, s.state.View(), false)
		if err != nil {
			return nil, err
		}

		s.state.rcc = RCC
	}

	block, _ := s.state.rcc.HighestRoundBlock()
	if block == nil {
		return s.BuildProposal(s.state.Sequence()), nil
	}

	return block, nil
}
