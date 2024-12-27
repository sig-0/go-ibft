package sequencer

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/sig-0/go-ibft/message"
)

type CommitSeal struct {
	From, Seal []byte
}

type SequenceResult struct {
	Proposal []byte
	Seals    []CommitSeal
	Round    uint64
}

// Sequencer is the consensus actor's (Validator) block finalization process. Whenever the network moves to a
// new sequence, all actors run their Sequencer processes to reach consensus on some proposal. Sequences consist of
// rounds in which a chosen actor (Proposer) suggests their own proposal to the network. The Sequencer makes sure
// that consensus is (eventually) reached, moving to higher rounds in case the network cannot agree on some proposal.
// Given its simple API method Finalize, Sequencer is designed to work alongside a syncing protocol
type Sequencer struct {
	validator      Validator
	validatorSet   ValidatorSet
	transport      message.Transport
	feed           message.Feed
	keccak         message.Keccak
	sig            message.SignatureVerifier
	state          state
	wg             sync.WaitGroup
	round0Duration time.Duration
}

// NewSequencer returns a Sequencer object for the provided validator
func NewSequencer(cfg Config) *Sequencer {
	return &Sequencer{
		validator:      cfg.Validator,
		validatorSet:   cfg.ValidatorSet,
		transport:      cfg.Transport,
		feed:           cfg.Feed,
		keccak:         cfg.Keccak,
		sig:            cfg.SignatureVerifier,
		round0Duration: cfg.Round0Duration,
	}
}

// Finalize runs the block finalization loop. This method returns a non-nil value only if consensus
// is reached for the provided sequence. Otherwise, it runs forever until cancelled by the caller
func (s *Sequencer) Finalize(ctx context.Context, sequence uint64) *SequenceResult {
	s.state.init(sequence)

	c := make(chan *SequenceResult, 1)
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
func (s *Sequencer) finalize(ctx context.Context) *SequenceResult {
	for {
		ctxRound, cancelRound := context.WithCancel(ctx)
		teardown := func() {
			cancelRound()
			s.wg.Wait()
		}

		select {
		case _, ok := <-s.startRoundTimer(ctxRound, s.state.round):
			teardown()
			if !ok {
				return nil
			}

			s.state.moveToNextRound()
			s.sendMsgRoundChange()

		case rcc, ok := <-s.awaitHigherRoundRCC(ctxRound, s.state.round):
			teardown()
			if !ok {
				return nil
			}

			s.state.acceptRCC(rcc)

		case proposal, ok := <-s.awaitHigherRoundProposal(ctxRound, s.state.round):
			teardown()
			if !ok {
				return nil
			}

			s.state.acceptProposal(proposal)
			s.sendMsgPrepare()

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
func (s *Sequencer) startRoundTimer(ctx context.Context, round uint64) <-chan struct{} {
	s.wg.Add(1)

	c := make(chan struct{}, 1)
	go func() {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		roundTimer := s.getRoundTimer(round)

		select {
		case <-ctx.Done():
			roundTimer.Stop()
		case <-roundTimer.C:
			c <- struct{}{}
		}
	}()

	return c
}

// awaitHigherRoundProposal listens for proposal messages from rounds higher than the current
func (s *Sequencer) awaitHigherRoundProposal(ctx context.Context, currentRound uint64) <-chan *message.MsgProposal {
	s.wg.Add(1)

	c := make(chan *message.MsgProposal, 1)
	go func() {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		proposal, err := s.awaitProposal(ctx, currentRound, true)
		if err != nil {
			return
		}

		c <- proposal
	}()

	return c
}

// awaitHigherRoundRCC listens for round change certificates from rounds higher than the current
func (s *Sequencer) awaitHigherRoundRCC(
	ctx context.Context,
	currentRound uint64,
) <-chan *message.RoundChangeCertificate {
	s.wg.Add(1)

	c := make(chan *message.RoundChangeCertificate, 1)
	go func() {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		rcc, err := s.awaitRCC(ctx, currentRound, true)
		if err != nil {
			return
		}

		c <- rcc
	}()

	return c
}

// awaitFinalizedBlockInCurrentRound starts the block finalization algorithm for the current round
func (s *Sequencer) awaitFinalizedBlockInCurrentRound(ctx context.Context) <-chan *SequenceResult {
	s.wg.Add(1)

	c := make(chan *SequenceResult, 1)
	go func() {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		if err := s.runRound(ctx); err != nil {
			return
		}

		c <- &SequenceResult{
			Round:    s.state.round,
			Proposal: s.state.proposal.ProposedBlock.Block,
			Seals:    s.state.seals,
		}
	}()

	return c
}

func (s *Sequencer) getRoundTimer(round uint64) *time.Timer {
	return time.NewTimer(s.round0Duration * time.Duration(math.Pow(2, float64(round))))
}

func (s *Sequencer) shouldPropose() bool {
	return s.validatorSet.IsProposer(s.validator.Address(), s.state.sequence, s.state.round)
}

func (s *Sequencer) buildProposal(ctx context.Context) ([]byte, error) {
	if s.state.round == 0 {
		return s.validator.BuildProposal(s.state.sequence), nil
	}

	if s.state.rcc == nil {
		// round jump triggered by round timer -> justify proposal with round change certificate
		RCC, err := s.awaitRCC(ctx, s.state.round, false)
		if err != nil {
			return nil, err
		}

		s.state.rcc = RCC
	}

	block, _ := s.state.rcc.HighestRoundBlock()
	if block == nil {
		return s.validator.BuildProposal(s.state.sequence), nil
	}

	return block, nil
}

func (s *Sequencer) runRound(ctx context.Context) error {
	if s.shouldPropose() {
		proposal, err := s.buildProposal(ctx)
		if err != nil {
			return err
		}

		s.sendMsgProposal(proposal)
	}

	if !s.state.isProposalAccepted() {
		proposal, err := s.awaitProposal(ctx, s.state.round, false)
		if err != nil {
			return err
		}

		s.state.acceptProposal(proposal)
		s.sendMsgPrepare()
	}

	prepares, err := s.awaitPrepareQuorum(ctx)
	if err != nil {
		return err
	}

	s.state.prepareCertificate(prepares)
	s.sendMsgCommit()

	commits, err := s.awaitCommitQuorum(ctx)
	if err != nil {
		return err
	}

	for _, commit := range commits {
		s.state.acceptSeal(commit.Info.Sender, commit.CommitSeal)
	}

	return nil
}
