package sequencer

import (
	"bytes"

	"github.com/sig-0/go-ibft"
	"github.com/sig-0/go-ibft/message/types"
)

func (s *Sequencer) sendMsgRoundChange(ctx Context) {
	msg := &types.MsgRoundChange{
		LatestPreparedProposedBlock: s.state.latestPB,
		LatestPreparedCertificate:   s.state.latestPC,
		Metadata: &types.MsgMetadata{
			View:   s.state.getView(),
			Sender: s.validator.ID(),
		},
	}

	msg.Metadata.Signature = s.validator.Sign(ctx.Keccak().Hash(msg.Payload()))

	ctx.Transport().MulticastRoundChange(msg)
}

func (s *Sequencer) awaitRCC(
	ctx Context,
	view *types.View,
	higherRounds bool,
) (*types.RoundChangeCertificate, error) {
	messages, err := s.awaitQuorumRoundChanges(ctx, view, higherRounds)
	if err != nil {
		return nil, err
	}

	return &types.RoundChangeCertificate{Messages: messages}, nil
}

func (s *Sequencer) awaitQuorumRoundChanges(
	ctx Context,
	view *types.View,
	higherRounds bool,
) ([]*types.MsgRoundChange, error) {
	if higherRounds {
		view.Round++
	}

	sub, cancelSub := ctx.MessageFeed().RoundChangeMessages(view, higherRounds)
	defer cancelSub()

	cache := newMsgCache(func(msg *types.MsgRoundChange) bool {
		return s.isValidMsgRoundChange(msg, ctx.Quorum(), ctx.Keccak())
	})

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case notification := <-sub:
			cache = cache.add(notification.Unwrap())

			roundChanges := cache.get()
			if len(roundChanges) == 0 || !ctx.Quorum().HasQuorum(types.WrapMessages(roundChanges...)) {
				continue
			}

			return roundChanges, nil
		}
	}
}

func (s *Sequencer) isValidMsgRoundChange(
	msg *types.MsgRoundChange,
	quorum ibft.Quorum,
	keccak ibft.Keccak,
) bool {
	if !s.validatorSet.IsValidator(msg.Sender(), msg.Sequence()) {
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

	if !s.isValidPC(pc, msg, quorum) {
		return false
	}

	if !bytes.Equal(pc.ProposalMessage.BlockHash, keccak.Hash(pb.Bytes())) {
		return false
	}

	return true
}

func (s *Sequencer) isValidPC(
	pc *types.PreparedCertificate,
	msg *types.MsgRoundChange,
	quorum ibft.Quorum,
) bool {
	if pc.ProposalMessage == nil || pc.PrepareMessages == nil {
		return false
	}

	if pc.ProposalMessage.Sequence() != msg.Sequence() {
		return false
	}

	if pc.ProposalMessage.Round() >= msg.Round() {
		return false
	}

	var (
		sequence = pc.ProposalMessage.Sequence()
		round    = pc.ProposalMessage.Round()
	)

	if !s.validatorSet.IsProposer(pc.ProposalMessage.Sender(), sequence, round) {
		return false
	}

	senders := map[string]struct{}{string(pc.ProposalMessage.Sender()): {}}

	for _, msg := range pc.PrepareMessages {
		if msg.Sequence() != sequence {
			return false
		}

		if msg.Round() != round {
			return false
		}

		if !bytes.Equal(msg.BlockHash, pc.ProposalMessage.BlockHash) {
			return false
		}

		if !s.validatorSet.IsValidator(msg.Sender(), sequence) {
			return false
		}

		senders[string(msg.Sender())] = struct{}{}
	}

	if len(senders) != 1+len(pc.PrepareMessages) {
		return false
	}

	return quorum.HasQuorum(append(
		types.WrapMessages(pc.ProposalMessage),
		types.WrapMessages(pc.PrepareMessages...)...,
	))
}

func (s *Sequencer) isValidRCC(
	rcc *types.RoundChangeCertificate,
	proposal *types.MsgProposal,
	quorum ibft.Quorum,
) bool {
	if rcc == nil || len(rcc.Messages) == 0 {
		return false
	}

	var (
		sequence = proposal.Sequence()
		round    = proposal.Round()
	)

	uniqueSenders := make(map[string]struct{})

	for _, msg := range rcc.Messages {
		if msg.Sequence() != sequence {
			return false
		}

		if msg.Round() != round {
			return false
		}

		if !s.validatorSet.IsValidator(msg.Sender(), sequence) {
			return false
		}

		uniqueSenders[string(msg.Sender())] = struct{}{}
	}

	if len(uniqueSenders) != len(rcc.Messages) {
		return false
	}

	if !quorum.HasQuorum(types.WrapMessages(rcc.Messages...)) {
		return false
	}

	return true
}
