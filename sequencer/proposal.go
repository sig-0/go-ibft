package sequencer

import (
	"bytes"
	"context"

	"github.com/sig-0/go-ibft/message"
)

func (s *Sequencer) sendMsgProposal(block []byte) {
	pb := &message.ProposedBlock{
		Block: block,
		Round: s.state.round,
	}

	msg := &message.Proposal{
		Sequence:               s.state.sequence,
		Round:                  s.state.round,
		Sender:                 s.validator.Address(),
		ProposedBlock:          pb,
		BlockHash:              s.keccak(pb.Bytes()),
		RoundChangeCertificate: s.state.rcc,
	}

	msg.Signature = s.validator.Sign(msg.Payload())

	s.state.proposal = msg
	s.transport.MulticastProposal(msg)
}

func (s *Sequencer) awaitProposal(ctx context.Context, round uint64, higherRounds bool) (*message.Proposal, error) {
	if higherRounds {
		round++
	}

	sub, cancelSub := s.feed.ProposalMessages.Subscribe(s.state.sequence, round, higherRounds)
	defer cancelSub()

	cache := message.NewMsgCache(s.isValidMsgProposal)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case notification := <-sub:
			cache.Add(notification()...)

			proposals := cache.Get()
			if len(proposals) == 0 {
				continue
			}

			return proposals[0], nil
		}
	}
}

func (s *Sequencer) isValidMsgProposal(msg *message.Proposal) bool {
	// msg round and proposed block round match
	if msg.ProposedBlock.Round != msg.Round {
		return false
	}

	// sender is part of the validator set
	if bytes.Equal(msg.Sender, s.validator.Address()) {
		return false
	}

	// sender is the selected proposer
	if !s.verifier.IsProposer(msg.Sender, msg.Sequence, msg.Round) {
		return false
	}

	// block hash and keccak hash of proposed block match
	if !bytes.Equal(msg.BlockHash, s.keccak(msg.ProposedBlock.Bytes())) {
		return false
	}

	if msg.Round == 0 {
		return s.verifier.IsValidProposal(msg.ProposedBlock.Block, msg.Sequence)
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
		return s.verifier.IsValidProposal(msg.ProposedBlock.Block, msg.Sequence)
	}

	// reuse the proposed block from previous (highest) round
	pb := &message.ProposedBlock{
		Block: msg.ProposedBlock.Block,
		Round: round,
	}

	// block hash and a keccak hash of proposed block match
	return bytes.Equal(blockHash, s.keccak(pb.Bytes()))
}
