package sequencer

import (
	"bytes"
	"context"
	"encoding/hex"
	"log/slog"
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

// Sequencer is the consensus.go actor's (Validator) block finalization process. Whenever the network moves to a
// new sequence, all actors run their Sequencer processes to reach consensus.go on some proposal. Sequences consist of
// rounds in which a chosen actor (Proposer) suggests their own proposal to the network. The Sequencer makes sure
// that consensus.go is (eventually) reached, moving to higher rounds in case the network cannot agree on some proposal.
// Given its simple API method Finalize, Sequencer is designed to work alongside a syncing protocol
type Sequencer struct {
	validator      Validator
	consensus      Consensus
	transport      Transport
	wg             sync.WaitGroup
	log            *slog.Logger
	round0Duration time.Duration
}

// NewSequencer returns a Sequencer object for the provided validator
func NewSequencer(
	log *slog.Logger,
	c Consensus,
	v Validator,
	t Transport,
	round0Duration time.Duration,
) *Sequencer {
	return &Sequencer{
		log:            log,
		consensus:      c,
		validator:      v,
		transport:      t,
		round0Duration: round0Duration,
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

		s.log.Debug("running finalize", "sequence", sequence.Number, "round", sequence.Round)
		now := time.Now()

		select {
		case _, ok := <-s.startRoundTimer(ctxRound, sequence):
			teardown()
			if !ok {
				return nil
			}

			s.log.Debug("round timer expired", "round", sequence.Round, "elapsed", time.Since(now).Seconds())

			sequence.Seals = nil
			sequence.Proposal = nil
			sequence.Round++

			msg := s.buildRoundChangeMessage(sequence)
			messages.RoundChangeMessages.Add(msg)
			s.transport.MulticastRoundChange(ctx, msg)

		case rcc, ok := <-s.awaitHigherRoundRCC(ctxRound, sequence, messages):
			teardown()
			if !ok {
				return nil
			}

			round := rcc.Messages[0].Round
			s.log.Debug("received round change certificate from higher round", "rcc_messages", len(rcc.Messages), "rcc_round", round)

			sequence.Seals = nil
			sequence.Proposal = nil
			sequence.RCC = rcc
			sequence.Round = round

		case proposal, ok := <-s.awaitHigherRoundProposal(ctxRound, sequence, messages):
			teardown()
			if !ok {
				return nil
			}

			s.log.Debug("received proposal from higher round", "proposal_hash", hex.EncodeToString(proposal.BlockHash), "proposal_round", proposal.Round)

			sequence.Proposal = proposal
			sequence.Round = proposal.Round
			sequence.Seals = nil

			msg := s.buildPrepareMessage(sequence)
			s.transport.MulticastPrepare(ctx, msg)
			messages.PrepareMessages.Add(msg)

		case _, ok := <-s.awaitFinalizedBlockInCurrentRound(ctxRound, sequence, messages):
			teardown()
			if !ok {
				return nil
			}

			s.log.Debug("proposal finalized!", "round", sequence.Round, "proposal_hash", hex.EncodeToString(sequence.Proposal.BlockHash), "seals", len(sequence.Seals))

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

		timer := time.NewTimer(s.getRoundDuration(round))

		select {
		case <-ctx.Done():
			timer.Stop()
		case <-timer.C:
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

		proposal, err := s.consensus.AwaitFutureProposal(ctx, *seq, store)
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

		messages, err := s.consensus.AwaitFutureRoundChange(ctx, *seq, store)
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

func (s *Sequencer) getRoundDuration(round uint64) time.Duration {
	return s.round0Duration * time.Duration(math.Pow(2, float64(round)))
}

func (s *Sequencer) buildProposal(
	ctx context.Context,
	sequence *Sequence,
	store *message.Store,
) ([]byte, error) {
	if sequence.Round != 0 {
		if sequence.RCC == nil {
			// higher round proposals must include rcc
			messages, err := s.consensus.AwaitRoundChange(ctx, *sequence, store)
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

	proposal, err := s.validator.BuildProposal(ctx, sequence.Number)
	if err != nil {
		return nil, err
	}

	return proposal, nil
}

func (s *Sequencer) runRound(
	ctx context.Context,
	sequence *Sequence,
	store *message.Store,
) error {
	if noProposalYet := sequence.Proposal == nil; noProposalYet {
		proposer, err := s.consensus.GetProposer(ctx, sequence.Number, sequence.Round)
		if err != nil {
			return err
		}

		if shouldPropose := bytes.Equal(proposer, s.validator.Address()); shouldPropose {
			s.log.Debug("I am the proposer", "proposer", hex.EncodeToString(proposer))

			proposal, err := s.buildProposal(ctx, sequence, store)
			if err != nil {
				return err
			}

			msg := s.buildProposalMessage(proposal, sequence)
			sequence.Proposal = msg
			s.transport.MulticastProposal(ctx, msg)
			s.log.Debug("broadcasted proposal message", "proposal_hash", hex.EncodeToString(msg.BlockHash))
		} else {
			proposal, err := s.consensus.AwaitProposal(ctx, *sequence, store)
			if err != nil {
				return err
			}

			s.log.Debug("received proposal", "proposal_hash", hex.EncodeToString(proposal.BlockHash))

			sequence.Proposal = proposal
			sequence.Round = proposal.Round
			sequence.Seals = nil

			msg := s.buildPrepareMessage(sequence)
			store.PrepareMessages.Add(msg)
			s.transport.MulticastPrepare(ctx, msg)
			s.log.Debug("broadcasted prepare msg", "proposal_hash", hex.EncodeToString(proposal.BlockHash))
		}
	}

	prepares, err := s.consensus.AwaitPrepare(ctx, *sequence, store)
	if err != nil {
		return err
	}

	s.log.Debug("passed quorum of prepare messages", "round", sequence.Round, "prepares", len(prepares))

	pc := &message.PreparedCertificate{
		ProposalMessage: sequence.Proposal,
		PrepareMessages: prepares,
	}

	sequence.LatestPB = sequence.Proposal.ProposedBlock
	sequence.LatestPC = pc

	msg := s.buildCommitMessage(sequence)
	store.CommitMessages.Add(msg)
	s.transport.MulticastCommit(ctx, msg)
	s.log.Debug("broadcasted commit message", "seal", hex.EncodeToString(msg.CommitSeal))

	commits, err := s.consensus.AwaitCommit(ctx, *sequence, store)
	if err != nil {
		return err
	}

	s.log.Debug("passed quorum of commit messages", "round", sequence.Round, "commits", len(commits))

	for _, commit := range commits {
		sequence.Seals = append(sequence.Seals, CommitSeal{
			From: commit.Sender,
			Seal: commit.CommitSeal,
		})
	}

	return nil
}
