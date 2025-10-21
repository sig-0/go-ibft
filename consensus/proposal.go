package consensus

import (
	"bytes"
	"context"
	"sort"

	"github.com/sig-0/go-ibft/message"
	"github.com/sig-0/go-ibft/sequencer"
)

func (c Consensus) AwaitProposal(
	ctx context.Context,
	sequence sequencer.Sequence,
	store *message.Store,
) (*message.Proposal, error) {
	sub, cancel := store.ProposalMessages.Subscribe(sequence.Number)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrap := <-sub:
			messages := unwrap()
			for _, msg := range messages {
				//if _, ok := seen[string(msg.Signature)]; ok {
				//	continue
				//}
				if msg.Round != sequence.Round {
					continue // only interested in this round
				}

				if !c.isValidProposal(ctx, sequence, msg) {
					continue
				}

				// as soon as we hit a valid proposal message, we should return
				// It's unlikely and benign if a proposer decides to gossip 2 valid proposals
				return msg, nil
			}
		}
	}
}

func (c Consensus) AwaitFutureProposal(
	ctx context.Context,
	sequence sequencer.Sequence,
	store *message.Store,
) (*message.Proposal, error) {
	seen := make(map[string]struct{})
	proposalsInRounds := make(map[uint64]*message.Proposal)

	sub, cancel := store.ProposalMessages.Subscribe(sequence.Number)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrap := <-sub:
			// todo: rework (highest to lowest)
			messages := unwrap()
			for _, msg := range messages {
				if _, ok := seen[string(msg.Signature)]; ok {
					continue
				}

				seen[string(msg.Signature)] = struct{}{}

				if msg.Round <= sequence.Round {
					// only interested in higher rounds
					continue
				}

				if !c.isValidProposal(ctx, sequence, msg) {
					continue
				}

				proposalsInRounds[msg.Round] = msg

				// now we check in descending order
				rounds := make([]uint64, 0, len(messages))
				for round := range proposalsInRounds {
					rounds = append(rounds, round)
				}

				sort.SliceStable(rounds, func(i, j int) bool { return rounds[i] > rounds[j] })

				// take the proposal from the highest round
				highestRound := rounds[0]
				return proposalsInRounds[highestRound], nil
			}
		}
	}
}

func (c Consensus) isValidProposal(
	ctx context.Context,
	sequence sequencer.Sequence,
	msg *message.Proposal,
) bool {
	if msg.ProposedBlock.Round != msg.Round {
		return false
	}

	// this is ensured in the next check
	//if bytes.Equal(msg.Sender, s.validator.Address()) {
	//	return false
	//}

	proposer, err := c.vs.GetProposer(ctx, sequence.Number, sequence.Round)
	if err != nil {
		return false // todo: log?
	}

	if notTheProposer := !bytes.Equal(msg.Sender, proposer); notTheProposer {
		return false
	}

	// block hash must match hash of the proposed block
	if !bytes.Equal(msg.BlockHash, message.GetProposalHash(msg.ProposedBlock)) {
		return false
	}

	if msg.Round == 0 {
		return c.proposal.Verify(ctx, msg.Sequence, msg.ProposedBlock.Block) == nil
	}

	/* non zero round proposals */

	rcc := msg.RoundChangeCertificate
	if !c.isValidRCC(ctx, rcc, msg) {
		return false
	}

	trimmedRCC := &message.RoundChangeCertificate{}
	for _, msg := range rcc.Messages {
		pc := msg.LatestPreparedCertificate
		if pc == nil {
			continue
		}

		// any included prepared certificate must be valid
		if c.isValidPC(ctx, pc, msg) {
			trimmedRCC.Messages = append(trimmedRCC.Messages, msg)
		}
	}

	blockHash, round := trimmedRCC.HighestRoundBlockHash()
	if blockHash == nil {
		// there is no previously agreed upon block hash, build a new proposal
		return c.proposal.Verify(ctx, msg.Sequence, msg.ProposedBlock.Block) == nil
	}

	// reuse the proposed block from previous (highest) round
	pb := &message.ProposedBlock{
		Block: msg.ProposedBlock.Block,
		Round: round,
	}

	// block hash and a keccak hash of proposed block match
	return bytes.Equal(blockHash, message.GetProposalHash(pb))
}

func (c Consensus) isValidRCC(
	ctx context.Context,
	rcc *message.RoundChangeCertificate,
	proposal *message.Proposal,
) bool {
	// rcc must be included
	if rcc == nil || len(rcc.Messages) == 0 {
		return false
	}

	var (
		sequence      = proposal.Sequence
		round         = proposal.Round
		uniqueSenders = make(map[string]struct{})
	)

	for _, msg := range rcc.Messages {
		// round change msg sequence (round) and proposal msg sequence (round) must match
		if msg.Sequence != sequence || msg.Round != round {
			return false
		}

		// sender must be part of the validator set
		if _, ok := c.currentValidators[string(msg.Sender)]; !ok {
			return false
		}

		uniqueSenders[string(msg.Sender)] = struct{}{}
	}

	// all messages must be unique
	if len(uniqueSenders) != len(rcc.Messages) {
		return false
	}

	senders := make([][]byte, len(uniqueSenders))
	for sender, _ := range uniqueSenders {
		senders = append(senders, []byte(sender))
	}

	ok, err := c.vs.CheckQuorum(ctx, sequence, senders)
	if err != nil {
		return false // todo: log
	}

	return ok
}
