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

type Config struct {
	Validator      Validator
	ValidatorSet   ProposerAlgo
	Consensus      Consensus
	Transport      Transport
	Round0Duration time.Duration
}

// Sequencer is the consensus actor's (Validator) block finalization process. Whenever the network moves to a
// new sequence, all actors run their Sequencer processes to reach consensus on some proposal. Sequences consist of
// rounds in which a chosen actor (Proposer) suggests their own proposal to the network. The Sequencer makes sure
// that consensus is (eventually) reached, moving to higher rounds in case the network cannot agree on some proposal.
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

// Finalize runs the block finalization loop. This method returns a non-nil value only if consensus
// is reached for the provided sequence. Otherwise, it runs forever until cancelled by the caller
func (s *Sequencer) Finalize(ctx context.Context, sequenceNumber uint64, messages *message.Store) *SequenceResult {
	sequence := Sequence{sequence: sequenceNumber}

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

			sequence.seals = nil
			sequence.proposal = nil
			sequence.round++

			msg := s.buildRoundChangeMessage(sequence)
			messages.RoundChangeMessages.Add(msg) // add to self
			s.transport.MulticastRoundChange(msg)

		case rcc, ok := <-s.awaitHigherRoundRCC(ctxRound, sequence, messages):
			teardown()
			if !ok {
				return nil
			}

			sequence.seals = nil
			sequence.proposal = nil
			sequence.rcc = rcc
			sequence.round = rcc.Messages[0].Round

		case proposal, ok := <-s.awaitHigherRoundProposal(ctxRound, sequence, messages):
			teardown()
			if !ok {
				return nil
			}

			s.acceptProposal(sequence, proposal)

		case _, ok := <-s.awaitFinalizedBlockInCurrentRound(ctxRound, sequence, messages):
			teardown()
			if !ok {
				return nil
			}

			return &SequenceResult{
				Round:    sequence.round,
				Proposal: sequence.proposal.ProposedBlock.Block,
				Seals:    sequence.seals,
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
	}(sequence.round)

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
	if sequence.round != 0 {
		if sequence.rcc == nil {
			// higher round proposals must include rcc
			messages, err := s.consensus.AwaitRoundChange(ctx, *sequence, store, false)
			if err != nil {
				return nil, err
			}

			sequence.rcc = &message.RoundChangeCertificate{Messages: messages}
		}

		block, _ := sequence.rcc.HighestRoundBlock()
		if block != nil {
			return block, nil
		}
	}

	return s.validator.BuildProposal(sequence.sequence), nil
}

func (s *Sequencer) runRound(
	ctx context.Context,
	sequence *Sequence,
	store *message.Store,
) error {
	if noProposalYet := sequence.proposal == nil; noProposalYet {
		proposer, err := s.validatorSet.GetProposer(ctx, sequence.sequence, sequence.round)
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

			s.acceptProposal(sequence, proposal)
		}
	}

	prepares, err := s.consensus.AwaitPrepare(ctx, *sequence, store)
	if err != nil {
		return err
	}

	s.acceptPrepare(sequence, prepares)

	commits, err := s.consensus.AwaitCommit(ctx, *sequence, store)
	if err != nil {
		return err
	}

	for _, commit := range commits {
		sequence.seals = append(sequence.seals, CommitSeal{
			From: commit.Sender,
			Seal: commit.CommitSeal,
		})
	}

	return nil
}

func (s *Sequencer) acceptProposal(sequence *Sequence, msg *message.Proposal) {
	sequence.proposal = msg
	sequence.round = msg.Round
	sequence.seals = nil

	s.transport.MulticastPrepare(s.buildPrepareMessage(sequence))
}

func (s *Sequencer) acceptPrepare(sequence *Sequence, msgs []*message.Prepare) {
	pc := &message.PreparedCertificate{
		ProposalMessage: sequence.proposal,
		PrepareMessages: msgs,
	}

	sequence.latestPB = sequence.proposal.ProposedBlock
	sequence.latestPC = pc

	s.transport.MulticastCommit(s.buildCommitMessage(sequence))
}
