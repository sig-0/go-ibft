package sequencer

import (
	"bytes"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

func (s *Sequencer) sendMsgProposal(ctx Context, block []byte) {
	pb := &types.ProposedBlock{
		Block: block,
		Round: s.state.Round(),
	}

	msg := &types.MsgProposal{
		From:                   s.ID(),
		View:                   s.state.View(),
		ProposedBlock:          pb,
		BlockHash:              ctx.Keccak().Hash(pb.Bytes()),
		RoundChangeCertificate: s.state.rcc,
	}

	msg.Signature = s.Sign(ctx.Keccak().Hash(msg.Payload()))

	s.state.proposal = msg

	ctx.MessageTransport().Proposal.Multicast(msg)
}

func (s *Sequencer) awaitCurrentRoundProposal(ctx Context) error {
	proposal, err := s.awaitProposal(ctx, s.state.view, false)
	if err != nil {
		return err
	}

	s.state.AcceptProposal(proposal)

	return nil
}

func (s *Sequencer) awaitProposal(ctx Context, view *types.View, higherRounds bool) (*types.MsgProposal, error) {
	if higherRounds {
		view.Round++
	}

	sub, cancelSub := ctx.MessageFeed().ProposalMessages(view, higherRounds)
	defer cancelSub()

	isValidMsg := func(msg *types.MsgProposal) bool {
		if !s.IsValidSignature(msg.GetSender(), ctx.Keccak().Hash(msg.Payload()), msg.GetSignature()) {
			return false
		}

		return s.isValidMsgProposal(msg, ctx.Quorum(), ctx.Keccak())
	}
	cache := newMsgCache(isValidMsg)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case notification := <-sub:
			cache = cache.Add(notification.Unwrap())

			proposals := cache.Get()
			if len(proposals) == 0 {
				continue
			}

			return proposals[0], nil
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
