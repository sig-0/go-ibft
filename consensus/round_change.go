package consensus

import (
	"bytes"
	"context"
	"sort"

	"github.com/sig-0/go-ibft/message"
	"github.com/sig-0/go-ibft/sequencer"
)

func (c Consensus) AwaitRoundChange(
	ctx context.Context,
	sequence sequencer.Sequence,
	store *message.Store,
) ([]*message.RoundChange, error) {
	sub, cancel := store.RoundChangeMessages.Subscribe(sequence.Number)
	defer cancel()

	seen := make(map[string]struct{})
	valid := make([]*message.RoundChange, 0)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrap := <-sub:
			messages := unwrap()
			for _, msg := range messages {
				if msg.Round != sequence.Round {
					continue // only interested in this particular round
				}

				if _, ok := seen[string(msg.Signature)]; ok {
					continue
				}

				if !c.isValidRoundChange(ctx, sequence, msg) {
					continue
				}

				valid = append(valid, msg)
			}

			senders := make([][]byte, 0, len(valid))
			for _, msg := range valid {
				senders = append(senders, msg.Sender)
			}

			ok, err := c.vs.CheckQuorum(ctx, sequence.Number, senders)
			if err != nil || !ok {
				continue
			}

			return valid, nil
		}
	}
}

func (c Consensus) AwaitFutureRoundChange(
	ctx context.Context,
	seq sequencer.Sequence,
	store *message.Store,
) ([]*message.RoundChange, error) {
	var (
		sequence    = seq.Number
		round       = seq.Round
		seen        = make(map[string]struct{})
		rccByRounds = make(map[uint64]*message.RoundChangeCertificate)
	)

	sub, cancel := store.RoundChangeMessages.Subscribe(seq.Number)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrap := <-sub:
			messages := unwrap()
			for _, msg := range messages {
				if msg.Round <= round {
					// only interested in higher rounds
					continue
				}

				if _, ok := seen[string(msg.Signature)]; ok {
					continue
				}

				seen[string(msg.Signature)] = struct{}{}

				if !c.isValidRoundChange(ctx, seq, msg) {
					continue
				}

				rcc := rccByRounds[msg.Round]
				if rcc == nil {
					rcc = &message.RoundChangeCertificate{}
					rccByRounds[msg.Round] = rcc
				}

				rcc.Messages = append(rcc.Messages, msg)
			}

			// now we check in descending order
			rounds := make([]uint64, 0, len(rccByRounds))
			for round := range rccByRounds {
				rounds = append(rounds, round)
			}

			sort.SliceStable(rounds, func(i, j int) bool { return rounds[i] > rounds[j] })

			getValidators := func(messages ...*message.RoundChange) [][]byte {
				validators := make([][]byte, 0, len(messages))
				for _, msg := range messages {
					validators = append(validators, msg.Sender)
				}

				return validators
			}

			// take the proposal from the highest round
			for _, round := range rounds {
				rcc := rccByRounds[round]
				ok, err := c.vs.CheckQuorum(ctx, sequence, getValidators(rcc.Messages...))
				if err != nil || !ok {
					continue
				}

				return rcc.Messages, nil
			}
		}
	}
}

func (c Consensus) isValidRoundChange(
	ctx context.Context,
	sequence sequencer.Sequence,
	msg *message.RoundChange,
) bool {
	//sender is part of the validator set
	if _, ok := c.currentValidators[string(msg.Sender)]; !ok {
		return false
	}

	var (
		pb = msg.LatestPreparedProposedBlock
		pc = msg.LatestPreparedCertificate
	)

	// if both pb and pc are missing, the message is valid
	if pb == nil && pc == nil {
		return true
	}

	// both pb and pc must be set
	if pb == nil || pc == nil {
		return false
	}

	if !c.isValidPC(ctx, pc, msg) {
		return false
	}

	// block hash in proposal message and keccak hash of proposed block match
	if !bytes.Equal(pc.ProposalMessage.BlockHash, message.GetProposalHash(pb)) {
		return false
	}

	return true
}

func (c Consensus) isValidPC(
	ctx context.Context,
	pc *message.PreparedCertificate,
	msg *message.RoundChange,
) bool {
	// both proposal message and prepare messages must be included
	if pc.ProposalMessage == nil || pc.PrepareMessages == nil {
		return false
	}

	var (
		sequence = pc.ProposalMessage.Sequence
		round    = pc.ProposalMessage.Round
	)

	if sequence != msg.Sequence {
		return false
	}

	if round >= msg.Round {
		return false
	}

	proposer, err := c.vs.GetProposer(ctx, sequence, round)
	if err != nil {
		return false // todo: log
	}

	// proposal sender in pc must be the selected proposer
	if notAProposer := !bytes.Equal(pc.ProposalMessage.Sender, proposer); notAProposer {
		return false
	}

	uniqueSenders := map[string]struct{}{
		string(pc.ProposalMessage.Sender): {}, // proposer
	}

	for _, msg := range pc.PrepareMessages {
		// prepare msg sequence (round) and proposal msg sequence (round) must match
		if msg.Sequence != sequence || msg.Round != round {
			return false
		}

		// prepare msg block hash and proposal msg block hash must match
		if !bytes.Equal(msg.BlockHash, pc.ProposalMessage.BlockHash) {
			return false
		}

		// prepare msg sender must be part of the validator set
		if _, ok := c.currentValidators[string(msg.Sender)]; !ok {
			return false
		}

		uniqueSenders[string(msg.Sender)] = struct{}{}
	}

	// 1 (proposer) + len(prepare) unique validators
	if len(uniqueSenders) != 1+len(pc.PrepareMessages) {
		return false
	}

	senders := make([][]byte, 0, len(uniqueSenders))
	for sender, _ := range uniqueSenders {
		senders = append(senders, []byte(sender))
	}

	// all messages in pc satisfy a quorum
	ok, err := c.vs.CheckQuorum(ctx, sequence, senders)
	if err != nil {
		return false // todo: log
	}

	return ok
}
