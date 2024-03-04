package sequencer

import (
	"github.com/madz-lab/go-ibft/message/types"
)

// viewState is a collection of consensus artifacts obtained at different
// stages of the currently running sequence.
type viewState struct {
	// the active sequence and round of this validator
	view *types.View

	// proposal that's being voted on
	proposal *types.MsgProposal

	// proposal that passed the PREPARE phase
	latestPB *types.ProposedBlock

	// proof that PREPARE was successful
	latestPC *types.PreparedCertificate

	// proof that ROUND CHANGE happened
	rcc *types.RoundChangeCertificate

	// proof that the proposal was finalized (passed COMMIT phase)
	seals []types.FinalizedSeal
}

func (s *viewState) Init(sequence uint64) {
	*s = viewState{view: &types.View{Sequence: sequence, Round: 0}}
}

func (s *viewState) View() *types.View {
	return &types.View{Sequence: s.view.Sequence, Round: s.view.Round}
}

func (s *viewState) Sequence() uint64 {
	return s.view.Sequence
}

func (s *viewState) Round() uint64 {
	return s.view.Round
}

func (s *viewState) ProposalAccepted() bool {
	return s.proposal != nil
}

func (s *viewState) AcceptedProposedBlock() *types.ProposedBlock {
	return s.proposal.ProposedBlock
}

func (s *viewState) AcceptedBlockHash() []byte {
	return s.proposal.BlockHash
}

func (s *viewState) MoveToNextRound() {
	s.view.Round++
	s.proposal = nil
	s.seals = s.seals[:0]
}

func (s *viewState) AcceptProposal(proposal *types.MsgProposal) {
	s.view.Round = proposal.View.Round
	s.proposal = proposal
	s.seals = s.seals[:0]
}

func (s *viewState) AcceptRCC(rcc *types.RoundChangeCertificate) {
	s.view.Round = rcc.Messages[0].View.Round
	s.rcc = rcc
	s.proposal = nil
	s.seals = s.seals[:0]
}

func (s *viewState) PrepareCertificate(prepares []*types.MsgPrepare) {
	s.latestPB, s.latestPC = s.AcceptedProposedBlock(), &types.PreparedCertificate{
		ProposalMessage: s.proposal,
		PrepareMessages: prepares,
	}
}

func (s *viewState) AcceptSeal(from, seal []byte) {
	s.seals = append(s.seals, types.FinalizedSeal{
		From:       from,
		CommitSeal: seal,
	})
}

func (s *viewState) FinalizedProposal() *types.FinalizedProposal {
	return &types.FinalizedProposal{
		Proposal: s.AcceptedProposedBlock().Block,
		Round:    s.Round(),
		Seals:    s.seals,
	}
}
