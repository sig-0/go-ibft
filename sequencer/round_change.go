package sequencer

import (
	"bytes"
	"context"

	"github.com/madz-lab/go-ibft/message/types"
)

func (s *Sequencer) awaitRCC(ctx context.Context, feed MessageFeed) (*types.RoundChangeCertificate, error) {
	messages, err := s.awaitQuorumRoundChangeMessages(ctx, feed)
	if err != nil {
		return nil, err
	}

	return &types.RoundChangeCertificate{Messages: messages}, nil
}

func (s *Sequencer) awaitFutureRCC(ctx context.Context, feed MessageFeed) (*types.RoundChangeCertificate, error) {
	messages, err := s.awaitQuorumFutureRoundChangeMessages(ctx, feed)
	if err != nil {
		return nil, err
	}

	return &types.RoundChangeCertificate{Messages: messages}, nil

}

func (s *Sequencer) awaitQuorumFutureRoundChangeMessages(ctx context.Context, feed MessageFeed) ([]*types.MsgRoundChange, error) {
	nextView := &types.View{
		Sequence: s.state.CurrentSequence(),
		Round:    s.state.CurrentRound() + 1,
	}

	sub, cancelSub := feed.SubscribeToRoundChangeMessages(nextView, true)
	defer cancelSub()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrapMessages := <-sub:
			validMessages := types.Filter(unwrapMessages(), s.isValidMsgRoundChange)
			if len(validMessages) == 0 {
				continue
			}

			if !s.quorum.HasQuorum(types.ToMsg(validMessages)) {
				continue
			}

			return validMessages, nil
		}
	}
}

func (s *Sequencer) awaitQuorumRoundChangeMessages(ctx context.Context, feed MessageFeed) ([]*types.MsgRoundChange, error) {
	sub, cancelSub := feed.SubscribeToRoundChangeMessages(s.state.currentView, false)
	defer cancelSub()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrapMessages := <-sub:
			validRoundChanges := types.Filter(unwrapMessages(), s.isValidMsgRoundChange)
			if len(validRoundChanges) == 0 {
				continue
			}

			if !s.quorum.HasQuorum(types.ToMsg(validRoundChanges)) {
				continue
			}

			return validRoundChanges, nil
		}
	}
}

func (s *Sequencer) isValidMsgRoundChange(msg *types.MsgRoundChange) bool {
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

	if !s.isValidPC(pc, msg.View.Sequence, msg.View.Round) {
		return false
	}

	if !bytes.Equal(pc.ProposalMessage.BlockHash, s.hash(pb)) {
		return false
	}

	return true
}

func (s *Sequencer) isValidPC(pc *types.PreparedCertificate, sequence, roundLimit uint64) bool {
	if pc.ProposalMessage == nil || pc.PrepareMessages == nil {
		return false
	}

	if pc.ProposalMessage.View.Sequence != sequence {
		return false
	}

	sequence, round := pc.ProposalMessage.View.Sequence, pc.ProposalMessage.View.Round
	for _, msg := range pc.PrepareMessages {
		if msg.View.Sequence != sequence {
			return false
		}

		if msg.View.Round != round {
			return false
		}
	}

	if pc.ProposalMessage.View.Round >= roundLimit {
		return false
	}

	for _, msg := range pc.PrepareMessages {
		if !bytes.Equal(msg.BlockHash, pc.ProposalMessage.BlockHash) {
			return false
		}
	}

	senders := map[string]struct{}{string(pc.ProposalMessage.From): {}}
	for _, msg := range pc.PrepareMessages {
		senders[string(msg.From)] = struct{}{}
	}

	if len(senders) != 1+len(pc.PrepareMessages) {
		return false
	}

	if !s.quorum.HasQuorum(append([]types.Msg{pc.ProposalMessage}, types.ToMsg(pc.PrepareMessages)...)) {
		return false
	}

	if !s.IsProposer(pc.ProposalMessage.From, pc.ProposalMessage.View.Sequence, pc.ProposalMessage.View.Round) {
		return false
	}

	for _, msg := range pc.PrepareMessages {
		if !s.IsValidator(msg.From, msg.View.Sequence) {
			return false
		}
	}

	return true
}

func (s *Sequencer) buildMsgRoundChange() *types.MsgRoundChange {
	msg := &types.MsgRoundChange{
		View:                        s.state.currentView,
		From:                        s.ID(),
		LatestPreparedProposedBlock: s.state.latestPreparedProposedBlock,
		LatestPreparedCertificate:   s.state.latestPreparedCertificate,
	}

	msg.Signature = s.Sign(msg.Payload())

	return msg
}

func (s *Sequencer) isValidRCC(rcc *types.RoundChangeCertificate, sequence, round uint64) bool {
	if rcc == nil {
		return false
	}

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

	l := len(rcc.Messages)
	q := s.quorum.HasQuorum(types.ToMsg(rcc.Messages))
	if !s.quorum.HasQuorum(types.ToMsg(rcc.Messages)) {
		return false
	}

	_, _ = l, q

	return true
}

func (s *Sequencer) filterInvalidPC(rcc *types.RoundChangeCertificate) *types.RoundChangeCertificate {
	valid := make([]*types.MsgRoundChange, 0, len(rcc.Messages))
	for _, msg := range rcc.Messages {
		pc := msg.LatestPreparedCertificate
		if pc == nil {
			continue
		}

		if s.isValidPC(pc, msg.View.Sequence, msg.View.Round) {
			valid = append(valid, msg)
		}
	}

	return &types.RoundChangeCertificate{Messages: valid}
}
