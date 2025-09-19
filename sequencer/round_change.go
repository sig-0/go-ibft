package sequencer

import (
	"context"

	"github.com/sig-0/go-ibft/message"
)

func (s *Sequencer) sendMsgRoundChange(sequence *Sequence) {
	msg := &message.RoundChange{
		Sequence:                    sequence.sequence,
		Round:                       sequence.round,
		Sender:                      s.validator.Address(),
		LatestPreparedProposedBlock: sequence.latestPB,
		LatestPreparedCertificate:   sequence.latestPC,
	}

	msg.Signature = s.validator.Sign(msg.Payload())

	s.feed.RoundChangeMessages.Add(msg) // add to self

	s.transport.MulticastRoundChange(msg)
}

func (s *Sequencer) awaitRCC(
	ctx context.Context,
	sequence *Sequence,
	higherRounds bool,
) (*message.RoundChangeCertificate, error) {
	round := sequence.round
	if higherRounds {
		round++
	}

	sub, cancelSub := s.feed.RoundChangeMessages.Subscribe(sequence.sequence, round, higherRounds)
	defer cancelSub()

	//cache := message.NewCache(s.isValidMsgRoundChange)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case notification := <-sub:
			//cache.Add(notification()...)

			messages, err := s.vrf.CheckRoundChange(ctx, sequence, notification())
			if err != nil {
				continue
				// todo: log
			}

			//roundChanges := cache.Get()
			//addresses := make([][]byte, 0, len(roundChanges))
			//for _, commit := range roundChanges {
			//	addresses = append(addresses, commit.GetSender())
			//}
			//
			//if len(roundChanges) == 0 || !s.verifier.HasQuorum(addresses, s.state.sequence) {
			//	continue
			//}

			return &message.RoundChangeCertificate{Messages: messages}, nil

		}
	}
}

//
//func (s *Sequencer) isValidMsgRoundChange(msg *message.RoundChange) bool {
//	// sender is part of the validator set
//	if !s.verifier.IsValidator(msg.Sender, msg.Sequence) {
//		return false
//	}
//
//	var (
//		pb = msg.LatestPreparedProposedBlock
//		pc = msg.LatestPreparedCertificate
//	)
//
//	// if both pb and pc are missing, the message is valid
//	if pb == nil && pc == nil {
//		return true
//	}
//
//	// both pb and pc must be set
//	if pb == nil || pc == nil {
//		return false
//	}
//
//	if !s.isValidPC(pc, msg) {
//		return false
//	}
//
//	// block hash in proposal message and keccak hash of proposed block match
//	if !bytes.Equal(pc.ProposalMessage.BlockHash, s.keccak(pb.Bytes())) {
//		return false
//	}
//
//	return true
//}
//
//func (s *Sequencer) isValidPC(pc *message.PreparedCertificate, msg *message.RoundChange) bool {
//	// both proposal message and prepare messages must be included
//	if pc.ProposalMessage == nil || pc.PrepareMessages == nil {
//		return false
//	}
//
//	var (
//		sequence = pc.ProposalMessage.Sequence
//		round    = pc.ProposalMessage.Round
//	)
//
//	// proposal sequence in pc and msg sequence must match
//	if sequence != msg.Sequence {
//		return false
//	}
//
//	// proposal round in pc must be higher than the round of the msg
//	if round >= msg.Round {
//		return false
//	}
//
//	// proposal sender in pc must be the selected proposer
//	if !s.verifier.IsProposer(pc.ProposalMessage.Sender, sequence, round) {
//		return false
//	}
//
//	uniqueSenders := map[string]struct{}{
//		string(pc.ProposalMessage.Sender): {}, // proposer
//	}
//
//	for _, msg := range pc.PrepareMessages {
//		// prepare msg sequence (round) and proposal msg sequence (round) must match
//		if msg.Sequence != sequence || msg.Round != round {
//			return false
//		}
//
//		// prepare msg block hash and proposal msg block hash must match
//		if !bytes.Equal(msg.BlockHash, pc.ProposalMessage.BlockHash) {
//			return false
//		}
//
//		// prepare msg sender must be part of the validator set
//		if !s.verifier.IsValidator(msg.Sender, sequence) {
//			return false
//		}
//
//		uniqueSenders[string(msg.Sender)] = struct{}{}
//	}
//
//	// 1 (proposer) + len(prepare) unique validators
//	if len(uniqueSenders) != 1+len(pc.PrepareMessages) {
//		return false
//	}
//
//	senders := make([][]byte, len(uniqueSenders))
//	for sender, _ := range uniqueSenders {
//		senders = append(senders, []byte(sender))
//	}
//
//	// all messages in pc satisfy a quorum
//	if !s.verifier.HasQuorum(senders, sequence) {
//		return false
//	}
//
//	return true
//}
//
//func (s *Sequencer) isValidRCC(rcc *message.RoundChangeCertificate, proposal *message.Proposal) bool {
//	// rcc must be included
//	if rcc == nil || len(rcc.Messages) == 0 {
//		return false
//	}
//
//	var (
//		sequence      = proposal.Sequence
//		round         = proposal.Round
//		uniqueSenders = make(map[string]struct{})
//	)
//
//	for _, msg := range rcc.Messages {
//		// round change msg sequence (round) and proposal msg sequence (round) must match
//		if msg.Sequence != sequence || msg.Round != round {
//			return false
//		}
//
//		// sender must be part of the validator set
//		if !s.verifier.IsValidator(msg.Sender, sequence) {
//			return false
//		}
//
//		uniqueSenders[string(msg.Sender)] = struct{}{}
//	}
//
//	// all messages must be unique
//	if len(uniqueSenders) != len(rcc.Messages) {
//		return false
//	}
//
//	senders := make([][]byte, len(uniqueSenders))
//	for sender, _ := range uniqueSenders {
//		senders = append(senders, []byte(sender))
//	}
//
//	// all messages in rcc satisfy a quorum
//	if !s.verifier.HasQuorum(senders, sequence) {
//		return false
//	}
//
//	return true
//}
