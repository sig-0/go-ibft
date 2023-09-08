package sequencer

import (
	"github.com/madz-lab/go-ibft/message/types"
)

type state struct {
	currentView                 *types.View
	acceptedProposal            *types.MsgProposal
	latestPreparedProposedBlock *types.ProposedBlock
	latestPreparedCertificate   *types.PreparedCertificate
	roundChangeCertificate      *types.RoundChangeCertificate
	seals                       []types.FinalizedSeal
}

func (s *state) CurrentSequence() uint64 {
	return s.currentView.Sequence
}

func (s *state) CurrentRound() uint64 {
	return s.currentView.Round
}

func (s *state) MoveToRound(round uint64) {
	s.seals = s.seals[:0]
	s.currentView.Round = round
	s.acceptedProposal = nil
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
