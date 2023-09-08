package sequencer

import (
	"github.com/madz-lab/go-ibft/message/types"
)

// state is a collection of consensus artifacts obtained at different
// stages of the currently running sequence.
type state struct {
	// the active sequence and round of this validator
	currentView *types.View

	// proposal accepted by the network
	acceptedProposal *types.MsgProposal

	// proposal that passed the PREPARE phase
	latestPreparedProposedBlock *types.ProposedBlock

	// proof that PREPARE was successful
	latestPreparedCertificate *types.PreparedCertificate

	// proof that ROUND CHANGE happened
	roundChangeCertificate *types.RoundChangeCertificate

	// proof that the proposal was finalized (passed COMMIT phase)
	seals []types.FinalizedSeal
}

func (s *state) CurrentView() *types.View {
	return &types.View{
		Sequence: s.currentView.Sequence,
		Round:    s.currentView.Round,
	}
}

func (s *state) CurrentSequence() uint64 {
	return s.currentView.Sequence
}

func (s *state) CurrentRound() uint64 {
	return s.currentView.Round
}

func (s *state) ProposalAccepted() bool {
	return s.acceptedProposal != nil
}

func (s *state) AcceptedProposedBlock() *types.ProposedBlock {
	return s.acceptedProposal.ProposedBlock
}

func (s *state) AcceptedBlockHash() []byte {
	return s.acceptedProposal.BlockHash
}

func (s *state) MoveToNextRound() {
	s.currentView.Round++
	s.acceptedProposal = nil
	s.seals = s.seals[:0]
}

func (s *state) AcceptProposal(proposal *types.MsgProposal) {
	s.currentView.Round = proposal.View.Round
	s.acceptedProposal = proposal
	s.seals = s.seals[:0]
}

func (s *state) AcceptRCC(rcc *types.RoundChangeCertificate) {
	s.currentView.Round = rcc.Messages[0].View.Round
	s.roundChangeCertificate = rcc
	s.acceptedProposal = nil
	s.seals = s.seals[:0]
}

func (s *state) PrepareCertificate(prepares []*types.MsgPrepare) {
	s.latestPreparedProposedBlock, s.latestPreparedCertificate = s.AcceptedProposedBlock(), &types.PreparedCertificate{
		ProposalMessage: s.acceptedProposal,
		PrepareMessages: prepares,
	}
}

func (s *state) AcceptSeal(from, seal []byte) {
	s.seals = append(s.seals, types.FinalizedSeal{
		From:       from,
		CommitSeal: seal,
	})
}

func (s *state) FinalizedBlock() *types.FinalizedBlock {
	return &types.FinalizedBlock{
		Block: s.AcceptedProposedBlock().Block,
		Round: s.CurrentRound(),
		Seals: s.seals,
	}
}
