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
		Round: s.state.round,
	}

	msg := &message.MsgProposal{
		Info: &message.MsgInfo{
			Sequence: s.state.sequence,
			Round:    s.state.round,
			Sender:   s.validator.Address(),
		},
		ProposedBlock:          pb,
		BlockHash:              s.keccak.Hash(pb.Bytes()),
		RoundChangeCertificate: s.state.rcc,
	}

	msg = message.SignMsg(msg, s.validator)
	s.state.proposal = msg
	s.transport.MulticastProposal(msg)
}

func (s *Sequencer) awaitProposal(ctx context.Context, higherRounds bool) (*message.MsgProposal, error) {
	round := s.state.round
	if higherRounds {
		round++
	}

	sub, cancelSub := s.feed.SubscribeProposal(s.state.sequence, round, higherRounds)
	defer cancelSub()

	cache := store.NewMsgCache(s.isValidMsgProposal)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case notification := <-sub:
			cache.Add(notification.Unwrap()...)

			proposals := cache.Get()
			if len(proposals) == 0 {
				continue
			}

			return proposals[0], nil
		}
	}
}

func (s *Sequencer) isValidMsgProposal(msg *message.MsgProposal) bool {
	// msg round and proposed block round match
	if msg.ProposedBlock.Round != msg.Info.Round {
		return false
	}

	// sender is part of the validator set
	if bytes.Equal(msg.Info.Sender, s.validator.Address()) {
		return false
	}

	// sender is the selected proposer
	if !s.validatorSet.IsProposer(msg.Info.Sender, msg.Info.Sequence, msg.Info.Round) {
		return false
	}

	// block hash and keccak hash of proposed block match
	if !bytes.Equal(msg.BlockHash, s.keccak.Hash(msg.ProposedBlock.Bytes())) {
		return false
	}

	if msg.Info.Round == 0 {
		return s.validator.IsValidProposal(msg.ProposedBlock.Block, msg.Info.Sequence)
	}

	/* non zero round proposals */

	rcc := msg.RoundChangeCertificate
	if !s.isValidRCC(rcc, msg) {
		return false
	}

	trimmedRCC := &message.RoundChangeCertificate{}
	for _, msg := range rcc.Messages {
		pc := msg.LatestPreparedCertificate
		if pc == nil {
			continue
		}

		// any included prepared certificate must be valid
		if s.isValidPC(pc, msg) {
			trimmedRCC.Messages = append(trimmedRCC.Messages, msg)
		}
	}

	blockHash, round := trimmedRCC.HighestRoundBlockHash()
	if blockHash == nil {
		// there is no previously agreed upon block hash, build a new proposal
		return s.validator.IsValidProposal(msg.ProposedBlock.Block, msg.Info.Sequence)
	}

	// reuse the proposed block from previous (highest) round
	pb := &message.ProposedBlock{
		Block: msg.ProposedBlock.Block,
		Round: round,
	}

	// block hash and a keccak hash of proposed block match
	if !bytes.Equal(blockHash, s.keccak.Hash(pb.Bytes())) {
		return false
	}

	return true
}
