package sequencer

import (
	"context"

	"github.com/sig-0/go-ibft/message"
)

func (s *Sequencer) buildProposalMessage(block []byte, sequence *Sequence) *message.Proposal {
	pb := &message.ProposedBlock{
		Block: block,
		Round: sequence.round,
	}

	msg := &message.Proposal{
		Sequence:               sequence.sequence,
		Round:                  sequence.round,
		Sender:                 s.validator.Address(),
		ProposedBlock:          pb,
		BlockHash:              s.keccak(pb.Bytes()),
		RoundChangeCertificate: sequence.rcc,
	}

	// todo: keccak this payload
	msg.Signature = s.validator.Sign(msg.Payload())

	sequence.proposal = msg
	
	return msg

}

func (s *Sequencer) awaitProposal(ctx context.Context, sequence *Sequence, higherRounds bool) (*message.Proposal, error) {
	round := sequence.round
	if higherRounds {
		round++
	}

	sub, cancelSub := s.feed.ProposalMessages.Subscribe(sequence.sequence, round, higherRounds)
	defer cancelSub()

	//cache := message.NewCache(s.isValidMsgProposal)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case notification := <-sub:
			//cache.Add(notification()...)
			msg, err := s.vrf.CheckProposal(ctx, *sequence, notification())
			if err != nil {
				// todo: log
				continue
			}

			return msg, nil

			//proposals := cache.Get()
			//if len(proposals) == 0 {
			//	continue
			//}
			//
			//return proposals[0], nil
		}
	}
}

//
//func (s *Sequencer) isValidMsgProposal(msg *message.Proposal) bool {
//	// msg round and proposed block round match
//	if msg.ProposedBlock.Round != msg.Round {
//		return false
//	}
//
//	// sender is part of the validator set
//	if bytes.Equal(msg.Sender, s.validator.Address()) {
//		return false
//	}
//
//	// sender is the selected proposer
//	if !s.verifier.IsProposer(msg.Sender, msg.Sequence, msg.Round) {
//		return false
//	}
//
//	// block hash and keccak hash of proposed block match
//	if !bytes.Equal(msg.BlockHash, s.keccak(msg.ProposedBlock.Bytes())) {
//		return false
//	}
//
//	if msg.Round == 0 {
//		return s.verifier.IsValidProposal(msg.ProposedBlock.Block, msg.Sequence)
//	}
//
//	/* non zero round proposals */
//
//	rcc := msg.RoundChangeCertificate
//	if !s.isValidRCC(rcc, msg) {
//		return false
//	}
//
//	trimmedRCC := &message.RoundChangeCertificate{}
//	for _, msg := range rcc.Messages {
//		pc := msg.LatestPreparedCertificate
//		if pc == nil {
//			continue
//		}
//
//		// any included prepared certificate must be valid
//		if s.isValidPC(pc, msg) {
//			trimmedRCC.Messages = append(trimmedRCC.Messages, msg)
//		}
//	}
//
//	blockHash, round := trimmedRCC.HighestRoundBlockHash()
//	if blockHash == nil {
//		// there is no previously agreed upon block hash, build a new proposal
//		return s.verifier.IsValidProposal(msg.ProposedBlock.Block, msg.Sequence)
//	}
//
//	// reuse the proposed block from previous (highest) round
//	pb := &message.ProposedBlock{
//		Block: msg.ProposedBlock.Block,
//		Round: round,
//	}
//
//	// block hash and a keccak hash of proposed block match
//	return bytes.Equal(blockHash, s.keccak(pb.Bytes()))
//}
