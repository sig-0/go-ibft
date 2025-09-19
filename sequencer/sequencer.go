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
	validator      Validator
	proposerAlgo   ProposerSelector
	vrf            Vrf
	transport      Transport
	feed           *message.Store
	keccak         KeccakFn
	wg             sync.WaitGroup
	round0Duration time.Duration
}

// NewSequencer returns a Sequencer object for the provided validator
func NewSequencer(cfg Config) *Sequencer {
	return &Sequencer{
		validator:      cfg.Validator,
		transport:      cfg.Transport,
		feed:           cfg.Feed,
		keccak:         cfg.Keccak,
		vrf:            cfg.Vrf,
		round0Duration: cfg.Round0Duration,
	}
}

// Finalize runs the block finalization loop. This method returns a non-nil value only if consensus
// is reached for the provided sequence. Otherwise, it runs forever until cancelled by the caller
func (s *Sequencer) Finalize(ctx context.Context, sequenceNumber uint64) *SequenceResult {
	sequence := Sequence{sequence: sequenceNumber}

	// todo: get validators

	c := make(chan *SequenceResult, 1)
	go func(seq *Sequence) {
		defer close(c)

		fb := s.finalize(ctx, seq)
		if fb == nil {
			return
		}

		c <- fb
	}(&sequence)

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
func (s *Sequencer) finalize(ctx context.Context, sequence *Sequence) *SequenceResult {
	for {
		ctxRound, cancelRound := context.WithCancel(ctx)
		teardown := func() {
			cancelRound()
			s.wg.Wait()
		}

		select {
		case _, ok := <-s.startRoundTimer(ctxRound, sequence):
			teardown()
			if !ok {
				return nil
			}

			sequence.moveToNextRound()
			s.sendMsgRoundChange(sequence)

		case rcc, ok := <-s.awaitHigherRoundRCC(ctxRound, sequence):
			teardown()
			if !ok {
				return nil
			}

			sequence.acceptRCC(rcc)

		case proposal, ok := <-s.awaitHigherRoundProposal(ctxRound, sequence):
			teardown()
			if !ok {
				return nil
			}

			sequence.acceptProposal(proposal)
			s.sendMsgPrepare(sequence)

		case fb, ok := <-s.awaitFinalizedBlockInCurrentRound(ctxRound, sequence):
			teardown()
			if !ok {
				return nil
			}

			return fb
		}
	}
}

// startRoundTimer starts the round timer of the current round
func (s *Sequencer) startRoundTimer(ctx context.Context, sequence *Sequence) <-chan struct{} {
	s.wg.Add(1)

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
	}(sequence.round)

	return c
}

// awaitHigherRoundProposal listens for proposal messages from rounds higher than the current
func (s *Sequencer) awaitHigherRoundProposal(ctx context.Context, sequence *Sequence) <-chan *message.Proposal {
	s.wg.Add(1)

	c := make(chan *message.Proposal, 1)

	go func(seq *Sequence) {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		proposal, err := s.awaitProposal(ctx, seq, true)
		if err != nil {
			return
		}

		c <- proposal
	}(sequence)

	return c
}

// awaitHigherRoundRCC listens for round change certificates from rounds higher than the current
func (s *Sequencer) awaitHigherRoundRCC(ctx context.Context, sequence *Sequence) <-chan *message.RoundChangeCertificate {
	s.wg.Add(1)

	c := make(chan *message.RoundChangeCertificate, 1)

	go func(seq *Sequence) {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		rcc, err := s.awaitRCC(ctx, seq, true)
		if err != nil {
			return
		}

		c <- rcc
	}(sequence)

	return c
}

// awaitFinalizedBlockInCurrentRound starts the block finalization algorithm for the current round
func (s *Sequencer) awaitFinalizedBlockInCurrentRound(ctx context.Context, sequence *Sequence) <-chan *SequenceResult {
	s.wg.Add(1)

	c := make(chan *SequenceResult, 1)
	go func(seq *Sequence) {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		if err := s.runRound(ctx, seq); err != nil {
			return
		}

		c <- &SequenceResult{
			Round:    sequence.round,
			Proposal: sequence.proposal.ProposedBlock.Block,
			Seals:    sequence.seals,
		}
	}(sequence)

	return c
}

func (s *Sequencer) getRoundTimer(round uint64) *time.Timer {
	return time.NewTimer(s.round0Duration * time.Duration(math.Pow(2, float64(round))))
}

// todo: should this be a critical error?
func (s *Sequencer) buildProposal(ctx context.Context, sequence *Sequence) ([]byte, error) {
	if sequence.round == 0 {
		return s.validator.BuildProposal(sequence.sequence), nil
	}

	if sequence.rcc == nil {
		// round jump triggered by round timer -> justify proposal with round change certificate
		RCC, err := s.awaitRCC(ctx, sequence, false)
		if err != nil {
			return nil, err
		}

		sequence.rcc = RCC
	}

	block, _ := sequence.rcc.HighestRoundBlock()
	if block == nil {
		return s.validator.BuildProposal(sequence.sequence), nil
	}

	return block, nil
}

func (s *Sequencer) runRound(ctx context.Context, sequence *Sequence) error {
	proposer, err := s.proposerAlgo.GetProposer(context.TODO(), sequence.sequence, sequence.round)
	if err != nil {
		panic(err)
	}

	if shouldPropose := bytes.Equal(proposer, s.validator.Address()); shouldPropose {
		proposal, err := s.buildProposal(ctx, sequence)
		if err != nil {
			return err
		}

		s.sendMsgProposal(proposal, sequence)
	}

	if !sequence.isProposalAccepted() {
		proposal, err := s.awaitProposal(ctx, sequence, false)
		if err != nil {
			return err
		}

		sequence.acceptProposal(proposal)
		s.sendMsgPrepare(sequence)
	}

	prepares, err := s.awaitPrepareQuorum(ctx, sequence)
	if err != nil {
		return err
	}

	sequence.prepareCertificate(prepares)
	s.sendMsgCommit(sequence)

	commits, err := s.awaitCommitQuorum(ctx, sequence)
	if err != nil {
		return err
	}

	for _, commit := range commits {
		sequence.acceptSeal(commit.Sender, commit.CommitSeal)
	}

	return nil
}
