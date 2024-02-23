package sequencer

import (
	"bytes"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

func (s *Sequencer) multicastRoundChange(ctx ibft.Context) {
	msg := &types.MsgRoundChange{
		View:                        s.state.currentView,
		From:                        s.ID(),
		LatestPreparedProposedBlock: s.state.latestPreparedProposedBlock,
		LatestPreparedCertificate:   s.state.latestPreparedCertificate,
	}

	msg.Signature = s.Sign(ctx.Keccak().Hash(msg.Payload()))

	ctx.Transport().Multicast(msg)
}

func (s *Sequencer) awaitRCC(
	ctx ibft.Context,
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
	ctx ibft.Context,
	view *types.View,
	higherRounds bool,
) ([]*types.MsgRoundChange, error) {
	if higherRounds {
		view.Round++
	}

	sub, cancelSub := ctx.Feed().RoundChangeMessages(view, higherRounds)
	defer cancelSub()

	cache := newMsgCache(func(msg *types.MsgRoundChange) bool {
		if !s.HasValidSignature(msg) {
			return false
		}

		return s.isValidMsgRoundChange(msg, ctx.Quorum(), ctx.Keccak())
	})

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case notification := <-sub:
			messages := notification.Unwrap()

			cache = cache.Add(messages)
			validRoundChanges := cache.Messages()

			if len(validRoundChanges) == 0 {
				continue
			}

			if !ctx.Quorum().HasQuorum(s.state.CurrentSequence(), ibft.WrapMessages(validRoundChanges)) {
				continue
			}

			return validRoundChanges, nil
		}
	}
}

func (s *Sequencer) isValidMsgRoundChange(
	msg *types.MsgRoundChange,
	quorum ibft.Quorum,
	keccak ibft.Keccak,
) bool {
	if !s.IsValidator(msg.From, msg.View.Sequence) {
		return false
	}

	var (
		pb = msg.LatestPreparedProposedBlock
		pc = msg.LatestPreparedCertificate
	)

	if pb == nil && pc == nil {
		return true
	}

	if pb == nil && pc != nil || pb != nil && pc == nil {
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

	if pc.ProposalMessage.View.Sequence != msg.View.Sequence {
		return false
	}

	if pc.ProposalMessage.View.Round >= msg.View.Round {
		return false
	}

	var (
		sequence = pc.ProposalMessage.View.Sequence
		round    = pc.ProposalMessage.View.Round
	)

	if !s.IsProposer(pc.ProposalMessage.From, sequence, round) {
		return false
	}

	senders := map[string]struct{}{string(pc.ProposalMessage.From): {}}

	for _, msg := range pc.PrepareMessages {
		if msg.View.Sequence != sequence {
			return false
		}

		if msg.View.Round != round {
			return false
		}

		if !bytes.Equal(msg.BlockHash, pc.ProposalMessage.BlockHash) {
			return false
		}

		if !s.IsValidator(msg.From, sequence) {
			return false
		}

		senders[string(msg.From)] = struct{}{}
	}

	if len(senders) != 1+len(pc.PrepareMessages) {
		return false
	}

	allMessages := []ibft.Message{pc.ProposalMessage}
	for _, msg := range pc.PrepareMessages {
		allMessages = append(allMessages, msg)
	}

	return quorum.HasQuorum(sequence, allMessages)
}

func (s *Sequencer) isValidRCC(
	rcc *types.RoundChangeCertificate,
	proposal *types.MsgProposal,
	quorum ibft.Quorum,
) bool {
	if rcc == nil {
		return false
	}

	var (
		sequence = proposal.View.Sequence
		round    = proposal.View.Round
	)

	if len(rcc.Messages) == 0 {
		return false
	}

	senders := make(map[string]struct{})

	for _, msg := range rcc.Messages {
		if msg.View.Sequence != sequence {
			return false
		}

		if msg.View.Round != round {
			return false
		}

		if !s.IsValidator(msg.From, sequence) {
			return false
		}

		senders[string(msg.From)] = struct{}{}
	}

	if len(senders) != len(rcc.Messages) {
		return false
	}

	if !quorum.HasQuorum(sequence, ibft.WrapMessages(rcc.Messages)) {
		return false
	}

	return true
}
