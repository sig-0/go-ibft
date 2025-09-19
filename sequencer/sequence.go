package sequencer

import (
	"github.com/sig-0/go-ibft/message"
)

// Sequence is a collection of consensus artifacts obtained by Sequencer during Finalize
type Sequence struct {
	// proposal that's being voted on
	proposal *message.Proposal

	// proposal that passed the PREPARE phase
	latestPB *message.ProposedBlock

	// proof that PREPARE was successful
	latestPC *message.PreparedCertificate

	// proof that ROUND CHANGE happened
	rcc *message.RoundChangeCertificate

	// proof that the proposal passed COMMIT phase
	seals []CommitSeal

	// currently running sequence
	sequence uint64

	// currently running round
	round uint64
}

func (s *Sequence) init(sequence uint64) {
	*s = Sequence{sequence: sequence}
}

func (s *Sequence) isProposalAccepted() bool {
	return s.proposal != nil
}

func (s *Sequence) acceptedBlockHash() []byte {
	return s.proposal.BlockHash
}

func (s *Sequence) moveToNextRound() {
	s.round++
	s.proposal = nil
	clear(s.seals)
}

func (s *Sequence) acceptProposal(proposal *message.Proposal) {
	s.proposal, s.round = proposal, proposal.Round
	clear(s.seals)
}

func (s *Sequence) acceptRCC(rcc *message.RoundChangeCertificate) {
	s.rcc, s.round, s.proposal = rcc, rcc.Messages[0].Round, nil
	clear(s.seals)
}

func (s *Sequence) prepareCertificate(prepares []*message.Prepare) {
	s.latestPB, s.latestPC = s.proposal.ProposedBlock, &message.PreparedCertificate{
		ProposalMessage: s.proposal,
		PrepareMessages: prepares,
	}
}

func (s *Sequence) acceptSeal(from, seal []byte) {
	s.seals = append(s.seals, CommitSeal{
		From: from,
		Seal: seal,
	})
}
