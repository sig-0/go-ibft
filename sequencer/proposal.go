package sequencer

import (
	"bytes"
	"context"

	"github.com/madz-lab/go-ibft/message/types"
)

func (s *Sequencer) awaitProposal(ctx context.Context, feed MessageFeed) error {
	if s.state.ProposalAccepted() {
		return nil
	}

	msg, err := s.awaitValidProposal(ctx, feed)
	if err != nil {
		return err
	}

	s.state.acceptedProposal = msg
	s.transport.Multicast(s.buildMsgPrepare())

	return nil
}

func (s *Sequencer) awaitFutureProposal(ctx context.Context, feed MessageFeed) (*types.MsgProposal, error) {
	nextView := &types.View{
		Sequence: s.state.CurrentSequence(),
		Round:    s.state.CurrentRound() + 1,
	}

	sub, cancelSub := feed.SubscribeToProposalMessages(nextView, true)
	defer cancelSub()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrapMessages := <-sub:
			validFutureProposals := types.Filter(unwrapMessages(), s.isValidMsgProposal)
			if len(validFutureProposals) == 0 {
				continue
			}

			// todo: proposer misbehavior
			//if len(validFutureProposals) > 1 {
			//}

			return validFutureProposals[0], nil
		}
	}
}

func (s *Sequencer) awaitValidProposal(ctx context.Context, feed MessageFeed) (*types.MsgProposal, error) {
	sub, cancelSub := feed.SubscribeToProposalMessages(s.state.currentView, false)
	defer cancelSub()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrapMessages := <-sub:
			validProposals := types.Filter(unwrapMessages(), s.isValidMsgProposal)
			if len(validProposals) == 0 {
				continue
			}

			//	todo proposer misbehavior
			//if len(validProposals) > 1 {
			//}

			return validProposals[0], nil
		}
	}
}

func (s *Sequencer) isValidMsgProposal(msg *types.MsgProposal) bool {
	if msg.ProposedBlock.Round != msg.View.Round {
		return false
	}

	if bytes.Equal(msg.From, s.ID()) {
		return false
	}

	if !s.IsProposer(msg.From, msg.View.Sequence, msg.View.Round) {
		return false
	}

	if !bytes.Equal(msg.BlockHash, s.hash(msg.ProposedBlock)) {
		return false
	}

	if msg.View.Round == 0 {
		if !s.IsValidBlock(msg.ProposedBlock.Block, 0) {
			return false
		}

		// all checks for round 0 proposal passed
		return true
	}

	rcc := msg.RoundChangeCertificate
	if !s.isValidRCC(rcc, msg.View.Sequence, msg.View.Round) {
		return false
	}

	rcc = s.filterInvalidPC(rcc)

	blockHash, round := rcc.HighestRoundBlockHash()
	if blockHash == nil {
		if !s.IsValidBlock(msg.ProposedBlock.Block, 0) {
			return false
		}

		return true
	}

	pb := &types.ProposedBlock{
		Block: msg.ProposedBlock.Block,
		Round: round,
	}

	if !bytes.Equal(blockHash, s.hash(pb)) {
		return false
	}

	return true
}

func (s *Sequencer) buildMsgProposal(block []byte) *types.MsgProposal {
	pb := &types.ProposedBlock{
		Block: block,
		Round: s.state.CurrentRound(),
	}

	msg := &types.MsgProposal{
		View:                   s.state.currentView,
		From:                   s.ID(),
		ProposedBlock:          pb,
		BlockHash:              s.hash(pb),
		RoundChangeCertificate: s.state.roundChangeCertificate,
	}

	msg.Signature = s.Sign(msg.Payload())

	return msg
}
