package sequencer

import (
	"github.com/sig-0/go-ibft/message"
)

// state is a collection of consensus artifacts obtained by Sequencer during Finalize
type state struct {
	// the active sequence and round of this validator
	view *message.View

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
}

func (s *state) init(sequence uint64) {
	*s = state{view: &message.View{Sequence: sequence, Round: 0}}
}

func (s *state) getView() *message.View {
	return &message.View{Sequence: s.view.Sequence, Round: s.view.Round}
}

func (s *state) getSequence() uint64 {
	return s.view.Sequence
}

func (s *state) getRound() uint64 {
	return s.view.Round
}

func (s *state) isProposalAccepted() bool {
	return s.proposal != nil
}

func (s *state) getProposedBlock() *message.ProposedBlock {
	return s.proposal.ProposedBlock
}

func (s *state) getProposedBlockHash() []byte {
	return s.proposal.BlockHash
}

func (s *state) moveToNextRound() {
	s.view.Round, s.proposal = s.view.Round+1, nil
	clear(s.seals)
}

func (s *state) acceptProposal(proposal *message.MsgProposal) {
	s.proposal, s.view.Round = proposal, proposal.Info.View.Round
	clear(s.seals)
}

func (s *state) acceptRCC(rcc *message.RoundChangeCertificate) {
	s.rcc, s.view.Round, s.proposal = rcc, rcc.Messages[0].Info.View.Round, nil
	clear(s.seals)
}

func (s *state) prepareCertificate(prepares []*message.MsgPrepare) {
	s.latestPB, s.latestPC = s.getProposedBlock(), &message.PreparedCertificate{
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
