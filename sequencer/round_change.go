package sequencer

import (
	"bytes"
	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

func (s *Sequencer) awaitRCC(ctx ibft.Context) (*types.RoundChangeCertificate, error) {
	messages, err := s.awaitQuorumRoundChangeMessages(ctx)
	if err != nil {
		return nil, err
	}

	return &types.RoundChangeCertificate{Messages: messages}, nil
}

func (s *Sequencer) awaitFutureRCC(ctx ibft.Context) (*types.RoundChangeCertificate, error) {
	messages, err := s.awaitQuorumFutureRoundChangeMessages(ctx)
	if err != nil {
		return nil, err
	}

	return &types.RoundChangeCertificate{Messages: messages}, nil
}

func (s *Sequencer) awaitQuorumFutureRoundChangeMessages(ctx ibft.Context) ([]*types.MsgRoundChange, error) {
	nextView := &types.View{
		Sequence: s.state.CurrentSequence(),
		Round:    s.state.CurrentRound() + 1,
	}

	cache := newMsgCache(func(msg *types.MsgRoundChange) bool {
		return s.isValidMsgRoundChange(msg, ctx.Quorum(), ctx.Keccak())
	})

	sub, cancelSub := ctx.Feed().RoundChange(nextView, true)
	defer cancelSub()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrapMessages := <-sub:
			cache = cache.Add(unwrapMessages())

			validMessages := cache.Messages()
			if len(validMessages) == 0 {
				continue
			}

			if !ctx.Quorum().HasQuorum(s.state.CurrentSequence(), types.ToMsg(validMessages)) {
				continue
			}

			return validMessages, nil
		}
	}
}

func (s *Sequencer) awaitQuorumRoundChangeMessages(ctx ibft.Context) ([]*types.MsgRoundChange, error) {
	cache := newMsgCache(func(msg *types.MsgRoundChange) bool {
		return s.isValidMsgRoundChange(msg, ctx.Quorum(), ctx.Keccak())
	})

	sub, cancelSub := ctx.Feed().RoundChange(s.state.currentView, false)
	defer cancelSub()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrapMessages := <-sub:
			cache = cache.Add(unwrapMessages())

			validRoundChanges := cache.Messages()
			if len(validRoundChanges) == 0 {
				continue
			}

			if !ctx.Quorum().HasQuorum(s.state.CurrentSequence(), types.ToMsg(validRoundChanges)) {
				continue
			}

			return validRoundChanges, nil
		}
	}
}

func (s *Sequencer) isValidMsgRoundChange(msg *types.MsgRoundChange, quorum ibft.Quorum, keccak ibft.Keccak) bool {
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

func (s *Sequencer) isValidPC(pc *types.PreparedCertificate, msg *types.MsgRoundChange, quorum ibft.Quorum) bool {
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

	if !quorum.HasQuorum(sequence, append([]types.Msg{pc.ProposalMessage}, types.ToMsg(pc.PrepareMessages)...)) {
		return false
	}

	return true
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

	if !quorum.HasQuorum(sequence, types.ToMsg(rcc.Messages)) {
		return false
	}

	return true
}
