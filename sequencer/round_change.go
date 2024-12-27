package sequencer

import (
	"bytes"
	"context"

	"github.com/sig-0/go-ibft/message"
	"github.com/sig-0/go-ibft/message/store"
)

func (s *Sequencer) sendMsgRoundChange() {
	msg := &message.MsgRoundChange{
		Info: &message.MsgInfo{
			Sequence: s.state.sequence,
			Round:    s.state.round,
			Sender:   s.validator.Address(),
		},
		LatestPreparedProposedBlock: s.state.latestPB,
		LatestPreparedCertificate:   s.state.latestPC,
	}

	s.transport.MulticastRoundChange(message.SignMsg(msg, s.validator))
}

func (s *Sequencer) awaitRCC(
	ctx context.Context,
	round uint64,
	higherRounds bool,
) (*message.RoundChangeCertificate, error) {
	if higherRounds {
		round++
	}

	sub, cancelSub := s.feed.SubscribeRoundChange(s.state.sequence, round, higherRounds)
	defer cancelSub()

	cache := store.NewMsgCache(s.isValidMsgRoundChange)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case notification := <-sub:
			cache.Add(notification.Unwrap()...)

			roundChanges := cache.Get()
			if len(roundChanges) == 0 || !s.validatorSet.HasQuorum(message.WrapMessages(roundChanges...)) {
				continue
			}

			return &message.RoundChangeCertificate{Messages: roundChanges}, nil
		}
	}
}

func (s *Sequencer) isValidMsgRoundChange(msg *message.MsgRoundChange) bool {
	// sender is part of the validator set
	if !s.validatorSet.IsValidator(msg.Info.Sender, msg.Info.Sequence) {
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

	if !s.isValidPC(pc, msg) {
		return false
	}

	// block hash in proposal message and keccak hash of proposed block match
	if !bytes.Equal(pc.ProposalMessage.BlockHash, s.keccak.Hash(pb.Bytes())) {
		return false
	}

	return true
}

func (s *Sequencer) isValidPC(pc *message.PreparedCertificate, msg *message.MsgRoundChange) bool {
	// both proposal message and prepare messages must be included
	if pc.ProposalMessage == nil || pc.PrepareMessages == nil {
		return false
	}

	var (
		sequence = pc.ProposalMessage.Info.Sequence
		round    = pc.ProposalMessage.Info.Round
	)

	// proposal sequence in pc and msg sequence must match
	if sequence != msg.Info.Sequence {
		return false
	}

	// proposal round in pc must be higher than the round of the msg
	if round >= msg.Info.Round {
		return false
	}

	// proposal sender in pc must be the selected proposer
	if !s.validatorSet.IsProposer(pc.ProposalMessage.Info.Sender, sequence, round) {
		return false
	}

	uniqueSenders := map[string]struct{}{
		string(pc.ProposalMessage.Info.Sender): {}, // proposer
	}

	for _, msg := range pc.PrepareMessages {
		// prepare msg sequence (round) and proposal msg sequence (round) must match
		if msg.Info.Sequence != sequence || msg.Info.Round != round {
			return false
		}

		// prepare msg block hash and proposal msg block hash must match
		if !bytes.Equal(msg.BlockHash, pc.ProposalMessage.BlockHash) {
			return false
		}

		// prepare msg sender must be part of the validator set
		if !s.validatorSet.IsValidator(msg.Info.Sender, sequence) {
			return false
		}

		uniqueSenders[string(msg.Info.Sender)] = struct{}{}
	}

	// 1 (proposer) + len(prepare) unique validators
	if len(uniqueSenders) != 1+len(pc.PrepareMessages) {
		return false
	}

	// all messages in pc satisfy a quorum
	if !s.validatorSet.HasQuorum(append(message.WrapMessages(pc.PrepareMessages...), pc.ProposalMessage)) {
		return false
	}

	return true
}

func (s *Sequencer) isValidRCC(rcc *message.RoundChangeCertificate, proposal *message.MsgProposal) bool {
	// rcc must be included
	if rcc == nil || len(rcc.Messages) == 0 {
		return false
	}

	var (
		sequence = proposal.Info.Sequence
		round    = proposal.Info.Round
	)

	uniqueSenders := make(map[string]struct{})

	for _, msg := range rcc.Messages {
		// round change msg sequence (round) and proposal msg sequence (round) must match
		if msg.Info.Sequence != sequence || msg.Info.Round != round {
			return false
		}

		// sender must be part of the validator set
		if !s.validatorSet.IsValidator(msg.Info.Sender, sequence) {
			return false
		}

		uniqueSenders[string(msg.Info.Sender)] = struct{}{}
	}

	// all messages must be unique
	if len(uniqueSenders) != len(rcc.Messages) {
		return false
	}

	// all messages in rcc satisfy a quorum
	if !s.validatorSet.HasQuorum(message.WrapMessages(rcc.Messages...)) {
		return false
	}

	return true
}
