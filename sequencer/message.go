package sequencer

import "github.com/sig-0/go-ibft/message"

func (s *Sequencer) buildProposalMessage(block []byte, sequence *Sequence) *message.Proposal {
	pb := &message.ProposedBlock{
		Block: block,
		Round: sequence.Round,
	}

	msg := &message.Proposal{
		Sequence:               sequence.Number,
		Round:                  sequence.Round,
		Sender:                 s.validator.Address(),
		ProposedBlock:          pb,
		BlockHash:              message.GetProposalHash(block, sequence.Round),
		RoundChangeCertificate: sequence.RCC,
	}

	// todo: keccak this payload
	msg.Signature = s.validator.Sign(msg.Payload())

	sequence.Proposal = msg

	return msg

}

func (s *Sequencer) buildPrepareMessage(sequence *Sequence) *message.Prepare {
	msg := &message.Prepare{
		Sequence:  sequence.Number,
		Round:     sequence.Round,
		Sender:    s.validator.Address(),
		BlockHash: sequence.Proposal.BlockHash,
	}

	// todo: keccak
	msg.Signature = s.validator.Sign(msg.Payload())

	return msg
}

func (s *Sequencer) buildCommitMessage(sequence *Sequence) *message.Commit {
	msg := &message.Commit{
		Sequence:   sequence.Number,
		Round:      sequence.Round,
		Sender:     s.validator.Address(),
		BlockHash:  sequence.Proposal.BlockHash,
		CommitSeal: s.validator.Sign(sequence.Proposal.BlockHash),
	}

	// todo: keccak
	msg.Signature = s.validator.Sign(msg.Payload())

	return msg
}

func (s *Sequencer) buildRoundChangeMessage(sequence *Sequence) *message.RoundChange {
	msg := &message.RoundChange{
		Sequence:                    sequence.Number,
		Round:                       sequence.Round,
		Sender:                      s.validator.Address(),
		LatestPreparedProposedBlock: sequence.LatestPB,
		LatestPreparedCertificate:   sequence.LatestPC,
	}

	msg.Signature = s.validator.Sign(msg.Payload())

	return msg
}
