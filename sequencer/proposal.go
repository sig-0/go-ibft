package sequencer

import (
	"bytes"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

func (s *Sequencer) multicastProposal(ctx ibft.Context, block []byte) {
	pb := &types.ProposedBlock{
		Block: block,
		Round: s.state.CurrentRound(),
	}

	msg := &types.MsgProposal{
		From:                   s.ID(),
		View:                   s.state.currentView,
		ProposedBlock:          pb,
		BlockHash:              ctx.Keccak().Hash(pb.Bytes()),
		RoundChangeCertificate: s.state.roundChangeCertificate,
	}

	msg.Signature = s.Sign(msg.Payload())

	s.state.acceptedProposal = msg

	ctx.Transport().Multicast(msg)
}

func (s *Sequencer) awaitCurrentRoundProposal(ctx ibft.Context) error {
	proposal, err := s.awaitProposal(ctx, s.state.currentView, false)
	if err != nil {
		return err
	}

	s.state.AcceptProposal(proposal)

	return nil
}

func (s *Sequencer) awaitProposal(ctx ibft.Context, view *types.View, higherRounds bool) (*types.MsgProposal, error) {
	if higherRounds {
		view.Round++
	}

	sub, cancelSub := ctx.Feed().ProposalMessages(view, higherRounds)
	defer cancelSub()

	cache := newMsgCache(func(msg *types.MsgProposal) bool {
		if !s.HasValidSignature(msg) {
			return false
		}

		return s.isValidMsgProposal(msg, ctx.Quorum(), ctx.Keccak())
	})

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case notification := <-sub:
			messages := notification.Unwrap()

			cache = cache.Add(messages)
			validProposals := cache.Messages()

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
		return s.IsValidProposal(msg.ProposedBlock.Block, msg.View.Sequence)
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
		return s.IsValidProposal(msg.ProposedBlock.Block, msg.View.Sequence)
	}

	pb := &types.ProposedBlock{
		Block: msg.ProposedBlock.Block,
		Round: round,
	}

	return bytes.Equal(blockHash, keccak.Hash(pb.Bytes()))
}
