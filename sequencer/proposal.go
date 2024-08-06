package sequencer

import (
	"bytes"
	"context"

	"github.com/sig-0/go-ibft/message"
	"github.com/sig-0/go-ibft/message/store"
)

func (s *Sequencer) sendMsgProposal(block []byte) {
	pb := &message.ProposedBlock{
		Block: block,
		Round: s.state.getRound(),
	}

	msg := &message.MsgProposal{
		Info: &message.MsgInfo{
			View:   s.state.getView(),
			Sender: s.validator.Address(),
		},
		ProposedBlock:          pb,
		BlockHash:              s.keccak.Hash(pb.Bytes()),
		RoundChangeCertificate: s.state.rcc,
	}

	msg = message.SignMsg(msg, s.validator)
	s.transport.MulticastProposal(msg)
}

func (s *Sequencer) awaitCurrentRoundProposal(ctx context.Context) error {
	proposal, err := s.awaitProposal(ctx, s.state.getView(), false)
	if err != nil {
		return err
	}

	s.state.acceptProposal(proposal)

	return nil
}

func (s *Sequencer) awaitProposal(ctx context.Context, view *message.View, higherRounds bool) (*message.MsgProposal, error) {
	if higherRounds {
		view.Round++
	}

	sub, cancelSub := s.feed.SubscribeProposal(view, higherRounds)
	defer cancelSub()

	cache := store.NewMsgCache(s.isValidMsgProposal)

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

func (s *Sequencer) isValidMsgProposal(msg *message.MsgProposal) bool {
	if msg.ProposedBlock.Round != msg.Info.View.Round {
		return false
	}

	if bytes.Equal(msg.Info.Sender, s.validator.Address()) {
		return false
	}

	if !s.validatorSet.IsProposer(msg.Info.Sender, msg.Info.View.Sequence, msg.Info.View.Round) {
		return false
	}

	if !bytes.Equal(msg.BlockHash, s.keccak.Hash(msg.ProposedBlock.Bytes())) {
		return false
	}

	if msg.Info.View.Round == 0 {
		return s.validator.IsValidProposal(msg.ProposedBlock.Block, msg.Info.View.Sequence)
	}

	rcc := msg.RoundChangeCertificate
	if !s.isValidRCC(rcc, msg) {
		return false
	}

	valid := make([]*message.MsgRoundChange, 0, len(rcc.Messages))

	for _, msg := range rcc.Messages {
		pc := msg.LatestPreparedCertificate
		if pc == nil {
			continue
		}

		if s.isValidPC(pc, msg) {
			valid = append(valid, msg)
		}
	}

	trimmedRCC := &message.RoundChangeCertificate{Messages: valid}

	blockHash, round := trimmedRCC.HighestRoundBlockHash()
	if blockHash == nil {
		return s.validator.IsValidProposal(msg.ProposedBlock.Block, msg.Info.View.Sequence)
	}

	pb := &message.ProposedBlock{
		Block: msg.ProposedBlock.Block,
		Round: round,
	}

	return bytes.Equal(blockHash, s.keccak.Hash(pb.Bytes()))
}
