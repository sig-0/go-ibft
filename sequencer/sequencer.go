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

// Sequence is a collection of consensus.go artifacts obtained by Sequencer during Finalize
type Sequence struct {
	// proposal that's being voted on
	Proposal *message.Proposal

	// proposal that passed the PREPARE phase
	LatestPB *message.ProposedBlock

	// proof that PREPARE was successful
	LatestPC *message.PreparedCertificate

	// proof that ROUND CHANGE happened
	RCC *message.RoundChangeCertificate

	// proof that the proposal passed COMMIT phase
	Seals []CommitSeal

	// currently running sequence
	Number uint64

	// currently running round
	Round uint64
}

type Config struct {
	Validator      Validator
	ValidatorSet   ProposerAlgo
	Consensus      Consensus
	Transport      Transport
	Round0Duration time.Duration
}

// Sequencer is the consensus.go actor's (Validator) block finalization process. Whenever the network moves to a
// new sequence, all actors run their Sequencer processes to reach consensus.go on some proposal. Sequences consist of
// rounds in which a chosen actor (Proposer) suggests their own proposal to the network. The Sequencer makes sure
// that consensus.go is (eventually) reached, moving to higher rounds in case the network cannot agree on some proposal.
// Given its simple API method Finalize, Sequencer is designed to work alongside a syncing protocol
type Sequencer struct {
	validator      Validator
	validatorSet   ProposerAlgo
	consensus      Consensus
	transport      Transport
	wg             sync.WaitGroup
	round0Duration time.Duration
}

// NewSequencer returns a Sequencer object for the provided validator
func NewSequencer(cfg Config) *Sequencer {
	return &Sequencer{
		validator:      cfg.Validator,
		transport:      cfg.Transport,
		consensus:      cfg.Consensus,
		round0Duration: cfg.Round0Duration,
		validatorSet:   cfg.ValidatorSet,
	}
}

// Finalize runs the block finalization loop. This method returns a non-nil value only if consensus.go
// is reached for the provided sequence. Otherwise, it runs forever until cancelled by the caller
func (s *Sequencer) Finalize(ctx context.Context, sequenceNumber uint64, messages *message.Store) *SequenceResult {
	sequence := Sequence{Number: sequenceNumber}

	c := make(chan *SequenceResult, 1)
	go func(seq *Sequence) {
		defer close(c)

		fb := s.finalize(ctx, seq, messages)
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
func (s *Sequencer) finalize(
	ctx context.Context,
	sequence *Sequence,
	messages *message.Store,
) *SequenceResult {
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

			sequence.Seals = nil
			sequence.Proposal = nil
			sequence.Round++

			msg := s.buildRoundChangeMessage(sequence)
			messages.RoundChangeMessages.Add(msg)
			s.transport.MulticastRoundChange(msg)

		case rcc, ok := <-s.awaitHigherRoundRCC(ctxRound, sequence, messages):
			teardown()
			if !ok {
				return nil
			}

			sequence.Seals = nil
			sequence.Proposal = nil
			sequence.RCC = rcc
			sequence.Round = rcc.Messages[0].Round

		case proposal, ok := <-s.awaitHigherRoundProposal(ctxRound, sequence, messages):
			teardown()
			if !ok {
				return nil
			}

			sequence.Proposal = proposal
			sequence.Round = proposal.Round
			sequence.Seals = nil

			msg := s.buildPrepareMessage(sequence)
			s.transport.MulticastPrepare(msg)
			messages.PrepareMessages.Add(msg)

		case _, ok := <-s.awaitFinalizedBlockInCurrentRound(ctxRound, sequence, messages):
			teardown()
			if !ok {
				return nil
			}

			return &SequenceResult{
				Round:    sequence.Round,
				Proposal: sequence.Proposal.ProposedBlock.Block,
				Seals:    sequence.Seals,
			}
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
	}(sequence.Round)

	return c
}

// awaitHigherRoundProposal listens for proposal messages from rounds higher than the current
func (s *Sequencer) awaitHigherRoundProposal(
	ctx context.Context,
	sequence *Sequence,
	store *message.Store,
) <-chan *message.Proposal {
	s.wg.Add(1)

	c := make(chan *message.Proposal, 1)

	go func(seq *Sequence) {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		proposal, err := s.consensus.AwaitProposal(ctx, *seq, store, true)
		if err != nil {
			return
		}

		c <- proposal
	}(sequence)

	return c
}

// awaitHigherRoundRCC listens for round change certificates from rounds higher than the current
func (s *Sequencer) awaitHigherRoundRCC(
	ctx context.Context,
	sequence *Sequence,
	store *message.Store,
) <-chan *message.RoundChangeCertificate {
	s.wg.Add(1)

	c := make(chan *message.RoundChangeCertificate, 1)

	go func(seq *Sequence, store *message.Store) {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		messages, err := s.consensus.AwaitRoundChange(ctx, *seq, store, true)
		if err != nil {
			return
		}

		c <- &message.RoundChangeCertificate{Messages: messages}
	}(sequence, store)

	return c
}

// awaitFinalizedBlockInCurrentRound starts the block finalization algorithm for the current round
func (s *Sequencer) awaitFinalizedBlockInCurrentRound(
	ctx context.Context,
	sequence *Sequence,
	store *message.Store,
) <-chan struct{} {
	s.wg.Add(1)

	c := make(chan struct{}, 1)
	go func(seq *Sequence) {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		if err := s.runRound(ctx, seq, store); err != nil {
			return
		}

		c <- struct{}{}
	}(sequence)

	return c
}

func (s *Sequencer) getRoundTimer(round uint64) *time.Timer {
	return time.NewTimer(s.round0Duration * time.Duration(math.Pow(2, float64(round))))
}

func (s *Sequencer) buildProposal(
	ctx context.Context,
	sequence *Sequence,
	store *message.Store,
) ([]byte, error) {
	if sequence.Round != 0 {
		if sequence.RCC == nil {
			// higher round proposals must include rcc
			messages, err := s.consensus.AwaitRoundChange(ctx, *sequence, store, false)
			if err != nil {
				return nil, err
			}

			sequence.RCC = &message.RoundChangeCertificate{Messages: messages}
		}

		block, _ := sequence.RCC.HighestRoundBlock()
		if block != nil {
			return block, nil
		}
	}

	return s.validator.BuildProposal(sequence.Number), nil
}

func (s *Sequencer) runRound(
	ctx context.Context,
	sequence *Sequence,
	store *message.Store,
) error {
	if noProposalYet := sequence.Proposal == nil; noProposalYet {
		proposer, err := s.validatorSet.GetProposer(ctx, sequence.Number, sequence.Round)
		if err != nil {
			return err
		}

		if shouldPropose := bytes.Equal(proposer, s.validator.Address()); shouldPropose {

			proposal, err := s.buildProposal(ctx, sequence, store)
			if err != nil {
				return err
			}

			s.transport.MulticastProposal(s.buildProposalMessage(proposal, sequence))
		} else {
			proposal, err := s.consensus.AwaitProposal(ctx, *sequence, store, false)
			if err != nil {
				return err
			}

			sequence.Proposal = proposal
			sequence.Round = proposal.Round
			sequence.Seals = nil

			msg := s.buildPrepareMessage(sequence)
			store.PrepareMessages.Add(msg)
			s.transport.MulticastPrepare(msg)
		}
	}

	prepares, err := s.consensus.AwaitPrepare(ctx, *sequence, store)
	if err != nil {
		return err
	}

	pc := &message.PreparedCertificate{
		ProposalMessage: sequence.Proposal,
		PrepareMessages: prepares,
	}

	sequence.LatestPB = sequence.Proposal.ProposedBlock
	sequence.LatestPC = pc

	msg := s.buildCommitMessage(sequence)
	//store.CommitMessages.Add(msg)
	s.transport.MulticastCommit(msg)

	commits, err := s.consensus.AwaitCommit(ctx, *sequence, store)
	if err != nil {
		return err
	}

	for _, commit := range commits {
		sequence.Seals = append(sequence.Seals, CommitSeal{
			From: commit.Sender,
			Seal: commit.CommitSeal,
		})
	}

	return nil
}
