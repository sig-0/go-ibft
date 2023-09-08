package sequencer

import (
	"bytes"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

func (s *Sequencer) awaitProposal(ctx ibft.Context) error {
	if s.state.ProposalAccepted() {
		return nil
	}

	proposal, err := s.awaitValidProposal(ctx)
	if err != nil {
		return err
	}

	s.state.acceptedProposal = proposal

	msg := &types.MsgPrepare{
		From:      s.ID(),
		View:      s.state.currentView,
		BlockHash: s.state.AcceptedBlockHash(),
	}

	msg.Signature = s.Sign(msg.Payload())

	ctx.Transport().Multicast(msg)

	return nil
}

func (s *Sequencer) awaitFutureProposal(ctx ibft.Context) (*types.MsgProposal, error) {
	nextView := &types.View{
		Sequence: s.state.CurrentSequence(),
		Round:    s.state.CurrentRound() + 1,
	}

	sub, cancelSub := ctx.Feed().Proposal(nextView, true)
	defer cancelSub()

	isValid := func(msg *types.MsgProposal) bool {
		return s.isValidMsgProposal(msg, ctx.Quorum(), ctx.Keccak())
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrapMessages := <-sub:
			validFutureProposals := types.Filter(unwrapMessages(), isValid)
			if len(validFutureProposals) == 0 {
				continue
			}

			return validFutureProposals[0], nil
		}
	}
}

func (s *Sequencer) awaitValidProposal(ctx ibft.Context) (*types.MsgProposal, error) {
	sub, cancelSub := ctx.Feed().Proposal(s.state.currentView, false)
	defer cancelSub()

	isValid := func(msg *types.MsgProposal) bool {
		return s.isValidMsgProposal(msg, ctx.Quorum(), ctx.Keccak())
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrapMessages := <-sub:
			validProposals := types.Filter(unwrapMessages(), isValid)
			if len(validProposals) == 0 {
				continue
			}

			return validProposals[0], nil
		}
	}
}

func (s *Sequencer) isValidMsgProposal(msg *types.MsgProposal, quorum ibft.Quorum, keccak ibft.Keccak) bool {
	if msg.ProposedBlock.Round != msg.View.Round {
		return false
	}

	if bytes.Equal(msg.From, s.ID()) {
		return false
	}

	if !s.IsProposer(msg.From, msg.View.Sequence, msg.View.Round) {
		return false
	}

	if !bytes.Equal(msg.BlockHash, keccak.Hash(msg.ProposedBlock.Bytes())) {
		return false
	}

	if msg.View.Round == 0 {
		return s.IsValidBlock(msg.ProposedBlock.Block, msg.View.Sequence)
	}

	rcc := msg.RoundChangeCertificate
	if !s.isValidRCC(rcc, msg, quorum) {
		return false
	}

	valid := make([]*types.MsgRoundChange, 0, len(rcc.Messages))

	for _, msg := range rcc.Messages {
		pc := msg.LatestPreparedCertificate
		if pc == nil {
			continue
		}

		if s.isValidPC(pc, msg, quorum) {
			valid = append(valid, msg)
		}
	}

	trimmedRCC := &types.RoundChangeCertificate{Messages: valid}

	blockHash, round := trimmedRCC.HighestRoundBlockHash()
	if blockHash == nil {
		return s.IsValidBlock(msg.ProposedBlock.Block, msg.View.Sequence)
	}

	pb := &types.ProposedBlock{
		Block: msg.ProposedBlock.Block,
		Round: round,
	}

	return bytes.Equal(blockHash, keccak.Hash(pb.Bytes()))
}
