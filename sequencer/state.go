package sequencer

import (
	"github.com/sig-0/go-ibft/message/types"
)

// state is a collection of consensus artifacts obtained by Sequencer during the finalization protocol
type state struct {
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

	// proof that the proposal passed COMMIT phase
	seals []types.FinalizedSeal
}

func (s *state) init(sequence uint64) {
	*s = state{view: &types.View{Sequence: sequence, Round: 0}}
}

func (s *state) getView() *types.View {
	return &types.View{Sequence: s.view.Sequence, Round: s.view.Round}
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

func (s *state) getProposedBlock() *types.ProposedBlock {
	return s.proposal.ProposedBlock
}

func (s *state) getProposedBlockHash() []byte {
	return s.proposal.BlockHash
}

func (s *state) moveToNextRound() {
	s.view.Round, s.proposal = s.view.Round+1, nil
	clear(s.seals)
}

func (s *state) acceptProposal(proposal *types.MsgProposal) {
	s.proposal, s.view.Round = proposal, proposal.Round()
	clear(s.seals)
}

func (s *state) acceptRCC(rcc *types.RoundChangeCertificate) {
	s.rcc, s.view.Round, s.proposal = rcc, rcc.Messages[0].Round(), nil
	clear(s.seals)
}

func (s *state) prepareCertificate(prepares []*types.MsgPrepare) {
	s.latestPB, s.latestPC = s.getProposedBlock(), &types.PreparedCertificate{
		ProposalMessage: s.proposal,
		PrepareMessages: prepares,
	}
}

func (s *state) acceptSeal(from, seal []byte) {
	s.seals = append(s.seals, types.FinalizedSeal{
		From:       from,
		CommitSeal: seal,
	})
}

func (s *state) getFinalizedProposal() *types.FinalizedProposal {
	return &types.FinalizedProposal{
		Proposal: s.getProposedBlock().Block,
		Round:    s.getRound(),
		Seals:    s.seals,
	}
}
