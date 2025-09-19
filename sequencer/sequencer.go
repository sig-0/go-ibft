package sequencer

import (
	"bytes"
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

// KeccakFn returns the KECCAK256 digest of arbitrary input
type KeccakFn func(data []byte) []byte

type Config struct {
	Validator      Validator
	ValidatorSet   Verifier
	Transport      Transport
	Feed           *message.Store
	Keccak         KeccakFn
	Round0Duration time.Duration
	Vrf            Vrf
}

// Sequencer is the consensus actor's (Validator) block finalization process. Whenever the network moves to a
// new sequence, all actors run their Sequencer processes to reach consensus on some proposal. Sequences consist of
// rounds in which a chosen actor (Proposer) suggests their own proposal to the network. The Sequencer makes sure
// that consensus is (eventually) reached, moving to higher rounds in case the network cannot agree on some proposal.
// Given its simple API method Finalize, Sequencer is designed to work alongside a syncing protocol
type Sequencer struct {
	validator Validator
	//verifier       Verifier
	proposerAlgo   ProposerSelector
	vrf            Vrf
	transport      Transport
	feed           *message.Store
	keccak         KeccakFn
	sequence       Sequence
	wg             sync.WaitGroup
	round0Duration time.Duration
}

// NewSequencer returns a Sequencer object for the provided validator
func NewSequencer(cfg Config) *Sequencer {
	return &Sequencer{
		validator: cfg.Validator,
		//verifier:       cfg.ValidatorSet,
		transport:      cfg.Transport,
		feed:           cfg.Feed,
		keccak:         cfg.Keccak,
		vrf:            cfg.Vrf,
		round0Duration: cfg.Round0Duration,
	}
}

// Finalize runs the block finalization loop. This method returns a non-nil value only if consensus
// is reached for the provided sequence. Otherwise, it runs forever until cancelled by the caller
func (s *Sequencer) Finalize(ctx context.Context, sequence uint64) *SequenceResult {
	s.sequence.init(sequence)

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
		case _, ok := <-s.startRoundTimer(ctxRound):
			teardown()
			if !ok {
				return nil
			}

			s.sequence.moveToNextRound()
			s.sendMsgRoundChange()

		case rcc, ok := <-s.awaitHigherRoundRCC(ctxRound):
			teardown()
			if !ok {
				return nil
			}

			s.sequence.acceptRCC(rcc)

		case proposal, ok := <-s.awaitHigherRoundProposal(ctxRound):
			teardown()
			if !ok {
				return nil
			}

			s.sequence.acceptProposal(proposal)
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
func (s *Sequencer) startRoundTimer(ctx context.Context) <-chan struct{} {
	s.wg.Add(1)

	round := s.sequence.round
	c := make(chan struct{}, 1)

	go func(round uint64) {
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
	}(round)

	return c
}

// awaitHigherRoundProposal listens for proposal messages from rounds higher than the current
func (s *Sequencer) awaitHigherRoundProposal(ctx context.Context) <-chan *message.Proposal {
	s.wg.Add(1)

	c := make(chan *message.Proposal, 1)
	round := s.sequence.round

	go func(round uint64) {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		proposal, err := s.awaitProposal(ctx, round, true)
		if err != nil {
			return
		}

		c <- proposal
	}(round)

	return c
}

// awaitHigherRoundRCC listens for round change certificates from rounds higher than the current
func (s *Sequencer) awaitHigherRoundRCC(ctx context.Context) <-chan *message.RoundChangeCertificate {
	s.wg.Add(1)

	c := make(chan *message.RoundChangeCertificate, 1)
	round := s.sequence.round

	go func(round uint64) {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		rcc, err := s.awaitRCC(ctx, round, true)
		if err != nil {
			return
		}

		c <- rcc
	}(round)

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
			Round:    s.sequence.round,
			Proposal: s.sequence.proposal.ProposedBlock.Block,
			Seals:    s.sequence.seals,
		}
	}()

	return c
}

func (s *Sequencer) getRoundTimer(round uint64) *time.Timer {
	return time.NewTimer(s.round0Duration * time.Duration(math.Pow(2, float64(round))))
}

func (s *Sequencer) shouldPropose() bool {
	proposer, err := s.proposerAlgo.GetProposer(context.TODO(), s.sequence.sequence, s.sequence.round)
	if err != nil {
		panic(err)
	}

	return bytes.Equal(proposer, s.validator.Address())
}

// todo: should this be a critical error?
func (s *Sequencer) buildProposal(ctx context.Context) ([]byte, error) {
	if s.sequence.round == 0 {
		return s.validator.BuildProposal(s.sequence.sequence), nil
	}

	if s.sequence.rcc == nil {
		// round jump triggered by round timer -> justify proposal with round change certificate
		RCC, err := s.awaitRCC(ctx, s.sequence.round, false)
		if err != nil {
			return nil, err
		}

		s.sequence.rcc = RCC
	}

	block, _ := s.sequence.rcc.HighestRoundBlock()
	if block == nil {
		return s.validator.BuildProposal(s.sequence.sequence), nil
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

	if !s.sequence.isProposalAccepted() {
		proposal, err := s.awaitProposal(ctx, s.sequence.round, false)
		if err != nil {
			return err
		}

		s.sequence.acceptProposal(proposal)
		s.sendMsgPrepare()
	}

	prepares, err := s.awaitPrepareQuorum(ctx)
	if err != nil {
		return err
	}

	s.sequence.prepareCertificate(prepares)
	s.sendMsgCommit()

	commits, err := s.awaitCommitQuorum(ctx)
	if err != nil {
		return err
	}

	for _, commit := range commits {
		s.sequence.acceptSeal(commit.Sender, commit.CommitSeal)
	}

	return nil
}
