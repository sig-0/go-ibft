package sequencer

import (
	"bytes"

	"github.com/sig-0/go-ibft"
	"github.com/sig-0/go-ibft/message/types"
)

func (s *Sequencer) sendMsgProposal(ctx Context, block []byte) {
	pb := &types.ProposedBlock{
		Block: block,
		Round: s.state.getRound(),
	}

	msg := &types.MsgProposal{
		ProposedBlock:          pb,
		BlockHash:              ctx.Keccak().Hash(pb.Bytes()),
		RoundChangeCertificate: s.state.rcc,
		Metadata: &types.MsgMetadata{
			Sender: s.ID(),
			View:   s.state.getView(),
		},
	}

	msg.Metadata.Signature = s.Sign(ctx.Keccak().Hash(msg.Payload()))

	s.state.proposal = msg

	ctx.Transport().MulticastProposal(msg)
}

func (s *Sequencer) awaitCurrentRoundProposal(ctx Context) error {
	proposal, err := s.awaitProposal(ctx, s.state.getView(), false)
	if err != nil {
		return err
	}

	s.state.acceptProposal(proposal)

	return nil
}

func (s *Sequencer) awaitProposal(ctx Context, view *types.View, higherRounds bool) (*types.MsgProposal, error) {
	if higherRounds {
		view.Round++
	}

	sub, cancelSub := ctx.MessageFeed().ProposalMessages(view, higherRounds)
	defer cancelSub()

	cache := newMsgCache(func(msg *types.MsgProposal) bool {
		return s.isValidMsgProposal(msg, ctx.Quorum(), ctx.Keccak())
	})

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case notification := <-sub:
			cache = cache.add(notification.Unwrap())

			proposals := cache.get()
			if len(proposals) == 0 {
				continue
			}

			return proposals[0], nil
		}
	}
}

func (s *Sequencer) isValidMsgProposal(msg *types.MsgProposal, quorum ibft.Quorum, keccak ibft.Keccak) bool {
	if msg.ProposedBlock.Round != msg.Round() {
		return false
	}

	if bytes.Equal(msg.Sender(), s.ID()) {
		return false
	}

	if !s.IsProposer(msg.Sender(), msg.Sequence(), msg.Round()) {
		return false
	}

	if !bytes.Equal(msg.BlockHash, keccak.Hash(msg.ProposedBlock.Bytes())) {
		return false
	}

	if msg.Round() == 0 {
		return s.IsValidProposal(msg.ProposedBlock.Block, msg.Sequence())
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
		return s.IsValidProposal(msg.ProposedBlock.Block, msg.Sequence())
	}

	pb := &types.ProposedBlock{
		Block: msg.ProposedBlock.Block,
		Round: round,
	}

	return bytes.Equal(blockHash, keccak.Hash(pb.Bytes()))
}
