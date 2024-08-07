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
			Sequence: s.state.getSequence(),
			Round:    s.state.getRound(),
			Sender:   s.validator.Address(),
		},
		LatestPreparedProposedBlock: s.state.latestPB,
		LatestPreparedCertificate:   s.state.latestPC,
	}

	msg = message.SignMsg(msg, s.validator)
	s.transport.MulticastRoundChange(msg)
}

func (s *Sequencer) awaitRCC(
	ctx context.Context,
	higherRounds bool,
) (*message.RoundChangeCertificate, error) {
	messages, err := s.awaitQuorumRoundChanges(ctx, higherRounds)
	if err != nil {
		return nil, err
	}

	return &message.RoundChangeCertificate{Messages: messages}, nil
}

func (s *Sequencer) awaitQuorumRoundChanges(ctx context.Context, higherRounds bool) ([]*message.MsgRoundChange, error) {
	round := s.state.getRound()
	if higherRounds {
		round++
	}

	sub, cancelSub := s.feed.SubscribeRoundChange(s.state.sequence, round, higherRounds)
	defer cancelSub()

	cache := store.NewMsgCache(func(msg *message.MsgRoundChange) bool {
		return s.isValidMsgRoundChange(msg)
	})

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case notification := <-sub:
			cache = cache.Add(notification.Unwrap())

			roundChanges := cache.Get()
			if len(roundChanges) == 0 || !s.validatorSet.HasQuorum(message.WrapMessages(roundChanges...)) {
				continue
			}

			return roundChanges, nil
		}
	}
}

func (s *Sequencer) isValidMsgRoundChange(msg *message.MsgRoundChange) bool {
	if !s.validatorSet.IsValidator(msg.Info.Sender, msg.Info.Sequence) {
		return false
	}

	var (
		pb = msg.LatestPreparedProposedBlock
		pc = msg.LatestPreparedCertificate
	)

	if pb == nil && pc == nil {
		return true
	}

	if pb == nil || pc == nil {
		return false
	}

	if !s.isValidPC(pc, msg) {
		return false
	}

	if !bytes.Equal(pc.ProposalMessage.BlockHash, s.keccak.Hash(pb.Bytes())) {
		return false
	}

	return true
}

func (s *Sequencer) isValidPC(
	pc *message.PreparedCertificate,
	msg *message.MsgRoundChange,
) bool {
	if pc.ProposalMessage == nil || pc.PrepareMessages == nil {
		return false
	}

	if pc.ProposalMessage.Info.Sequence != msg.Info.Sequence {
		return false
	}

	if pc.ProposalMessage.Info.Round >= msg.Info.Round {
		return false
	}

	var (
		sequence = pc.ProposalMessage.Info.Sequence
		round    = pc.ProposalMessage.Info.Round
	)

	if !s.validatorSet.IsProposer(pc.ProposalMessage.Info.Sender, sequence, round) {
		return false
	}

	senders := map[string]struct{}{string(pc.ProposalMessage.Info.Sender): {}}

	for _, msg := range pc.PrepareMessages {
		if msg.Info.Sequence != sequence {
			return false
		}

		if msg.Info.Round != round {
			return false
		}

		if !bytes.Equal(msg.BlockHash, pc.ProposalMessage.BlockHash) {
			return false
		}

		if !s.validatorSet.IsValidator(msg.Info.Sender, sequence) {
			return false
		}

		senders[string(msg.Info.Sender)] = struct{}{}
	}

	if len(senders) != 1+len(pc.PrepareMessages) {
		return false
	}

	if !s.validatorSet.HasQuorum(append(message.WrapMessages(pc.PrepareMessages...), pc.ProposalMessage)) {
		return false
	}

	return true
}

func (s *Sequencer) isValidRCC(
	rcc *message.RoundChangeCertificate,
	proposal *message.MsgProposal,
) bool {
	if rcc == nil || len(rcc.Messages) == 0 {
		return false
	}

	var (
		sequence = proposal.Info.Sequence
		round    = proposal.Info.Round
	)

	uniqueSenders := make(map[string]struct{})

	for _, msg := range rcc.Messages {
		if msg.Info.Sequence != sequence {
			return false
		}

		if msg.Info.Round != round {
			return false
		}

		if !s.validatorSet.IsValidator(msg.Info.Sender, sequence) {
			return false
		}

		uniqueSenders[string(msg.Info.Sender)] = struct{}{}
	}

	if len(uniqueSenders) != len(rcc.Messages) {
		return false
	}

	if !s.validatorSet.HasQuorum(message.WrapMessages(rcc.Messages...)) {
		return false
	}

	return true
}
