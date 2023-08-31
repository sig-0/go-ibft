package sequencer

import "github.com/madz-lab/go-ibft/message/types"

type mockValidator struct {
	idFn         func() []byte
	signFn       func([]byte) []byte
	buildBlockFn func() []byte
}

func (v mockValidator) ID() []byte {
	return v.idFn()
}

func (v mockValidator) Sign(bytes []byte) []byte {
	return v.signFn(bytes)
}

func (v mockValidator) BuildBlock() []byte {
	return v.buildBlockFn()
}

type mockVerifier struct {
	keccakFn       func([]byte) []byte
	isValidBlockFn func([]byte) bool
	isProposerFn   func(*types.View, []byte) bool
	recoverFromFn  func([]byte, []byte) []byte
	isValidatorFn  func([]byte, uint64) bool
}

func (v mockVerifier) Keccak(bytes []byte) []byte {
	return v.keccakFn(bytes)
}

func (v mockVerifier) IsValidBlock(bytes []byte) bool {
	return v.isValidBlockFn(bytes)
}

func (v mockVerifier) IsProposer(view *types.View, id []byte) bool {
	return v.isProposerFn(view, id)
}

func (v mockVerifier) RecoverFrom(data []byte, sig []byte) []byte {
	return v.recoverFromFn(data, sig)
}

func (v mockVerifier) IsValidator(id []byte, height uint64) bool {
	return v.isValidatorFn(id, height)
}

type mockMessageeFeed struct {
	proposalsByView    map[uint64]map[uint64][]*types.MsgProposal
	preparesByView     map[uint64]map[uint64][]*types.MsgPrepare
	commitsByView      map[uint64]map[uint64][]*types.MsgCommit
	roundChangesByView map[uint64]map[uint64][]*types.MsgRoundChange
	subProposalFn      func() []*types.MsgProposal
	subPrepareFn       func() []*types.MsgPrepare
	subCommitFn        func() []*types.MsgCommit
	subRoundChangeFn   func() []*types.MsgRoundChange
}

func (f mockMessageeFeed) SubscribeToProposalMessages(view *types.View, _ bool) (<-chan func() []*types.MsgProposal, func()) {
	c := make(chan func() []*types.MsgProposal, 1)
	c <- func() []*types.MsgProposal {
		return f.proposalsByView[view.Sequence][view.Round]
	}

	return c, func() {}
}

func (f mockMessageeFeed) SubscribeToPrepareMessages(view *types.View, _ bool) (<-chan func() []*types.MsgPrepare, func()) {
	c := make(chan func() []*types.MsgPrepare, 1)
	c <- func() []*types.MsgPrepare {
		return f.preparesByView[view.Sequence][view.Round]
	}

	return c, func() {}
}

func (f mockMessageeFeed) SubscribeToCommitMessages(view *types.View, _ bool) (<-chan func() []*types.MsgCommit, func()) {
	c := make(chan func() []*types.MsgCommit, 1)
	c <- func() []*types.MsgCommit {
		return f.commitsByView[view.Sequence][view.Round]
	}

	return c, func() {}
}

func (f mockMessageeFeed) SubscribeToRoundChangeMessages(view *types.View, _ bool) (<-chan func() []*types.MsgRoundChange, func()) {
	c := make(chan func() []*types.MsgRoundChange, 1)
	c <- func() []*types.MsgRoundChange {
		return f.roundChangesByView[view.Sequence][view.Round]
	}

	return c, func() {}
}

type mockQuorum struct {
	quorumPrepare     func(...*types.MsgPrepare) bool
	quorumCommit      func(...*types.MsgCommit) bool
	quorumRoundChange func(...*types.MsgRoundChange) bool
}

func (q mockQuorum) HasQuorumPrepareMessages(prepare ...*types.MsgPrepare) bool {
	return q.quorumPrepare(prepare...)
}

func (q mockQuorum) HasQuorumCommitMessages(commit ...*types.MsgCommit) bool {
	return q.quorumCommit(commit...)
}

func (q mockQuorum) HasQuorumRoundChangeMessages(roundchange ...*types.MsgRoundChange) bool {
	return q.quorumRoundChange(roundchange...)
}
