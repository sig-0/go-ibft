package sequencer

import (
	"github.com/sig-0/go-ibft/message"
)

// state is a collection of consensus artifacts obtained by Sequencer during Finalize
type state struct {
	// proposal that's being voted on
	proposal *message.MsgProposal

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

func (s *state) init(sequence uint64) {
	*s = state{sequence: sequence}
}

func (s *state) isProposalAccepted() bool {
	return s.proposal != nil
}

func (s *state) acceptedBlockHash() []byte {
	return s.proposal.BlockHash
}

func (s *state) moveToNextRound() {
	s.round++
	s.proposal = nil
	clear(s.seals)
}

func (s *state) acceptProposal(proposal *message.MsgProposal) {
	s.proposal, s.round = proposal, proposal.Info.Round
	clear(s.seals)
}

func (s *state) acceptRCC(rcc *message.RoundChangeCertificate) {
	s.rcc, s.round, s.proposal = rcc, rcc.Messages[0].Info.Round, nil
	clear(s.seals)
}

func (s *state) prepareCertificate(prepares []*message.MsgPrepare) {
	s.latestPB, s.latestPC = s.proposal.ProposedBlock, &message.PreparedCertificate{
		ProposalMessage: s.proposal,
		PrepareMessages: prepares,
	}
}

func (s *state) acceptSeal(from, seal []byte) {
	s.seals = append(s.seals, CommitSeal{
		From: from,
		Seal: seal,
	})
}
