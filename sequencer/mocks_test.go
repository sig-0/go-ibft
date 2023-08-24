package sequencer

import "github.com/madz-lab/go-ibft/message/types"

type mockValidator struct {
	recoverFromFn  func([]byte, []byte) []byte
	hashFn         func([]byte) []byte
	isValidBlockFn func([]byte) bool
	idFn           func() []byte
	isProposerFn   func(*types.View, []byte) bool
	signFn         func([]byte) []byte
	buildBlockFn   func() []byte
}

func (v mockValidator) RecoverFrom(data []byte, sig []byte) []byte {
	return v.recoverFromFn(data, sig)
}

func (v mockValidator) Hash(bytes []byte) []byte {
	return v.hashFn(bytes)
}

func (v mockValidator) IsValidBlock(bytes []byte) bool {
	return v.isValidBlockFn(bytes)
}

func (v mockValidator) ID() []byte {
	return v.idFn()
}

func (v mockValidator) IsProposer(view *types.View, id []byte) bool {
	return v.isProposerFn(view, id)
}

func (v mockValidator) Sign(bytes []byte) []byte {
	return v.signFn(bytes)
}

func (v mockValidator) BuildBlock() []byte {
	return v.buildBlockFn()
}

type mockMessageeFeed struct {
	subProposalFn    func() []*types.MsgProposal
	subPrepareFn     func() []*types.MsgPrepare
	subCommitFn      func() []*types.MsgCommit
	subRoundChangeFn func() []*types.MsgRoundChange
}

func (f mockMessageeFeed) SubscribeToProposalMessages(view *types.View, higherRounds bool) (<-chan func() []*types.MsgProposal, func()) {
	c := make(chan func() []*types.MsgProposal, 1)
	c <- f.subProposalFn

	return c, func() {}
}

func (f mockMessageeFeed) SubscribeToPrepareMessages(view *types.View, higherRounds bool) (<-chan func() []*types.MsgPrepare, func()) {
	c := make(chan func() []*types.MsgPrepare, 1)
	c <- f.subPrepareFn

	return c, func() {}
}

func (f mockMessageeFeed) SubscribeToCommitMessages(view *types.View, higherRounds bool) (<-chan func() []*types.MsgCommit, func()) {
	c := make(chan func() []*types.MsgCommit, 1)
	c <- f.subCommitFn

	return c, func() {}
}

func (f mockMessageeFeed) SubscribeToRoundChangeMessages(view *types.View, higherRounds bool) (<-chan func() []*types.MsgRoundChange, func()) {
	c := make(chan func() []*types.MsgRoundChange, 1)
	c <- f.subRoundChangeFn

	return c, func() {}
}

type mockQuorum struct {
	quorumPrepare func(...*types.MsgPrepare) bool
	quorumCommit  func(...*types.MsgCommit) bool
}

func (q mockQuorum) HasQuorumPrepareMessages(prepare ...*types.MsgPrepare) bool {
	return q.quorumPrepare(prepare...)
}

func (q mockQuorum) HasQuorumCommitMessages(commit ...*types.MsgCommit) bool {
	return q.quorumCommit(commit...)
}
