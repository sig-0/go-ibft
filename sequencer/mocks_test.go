package sequencer

import "github.com/madz-lab/go-ibft/message/types"

type msg interface {
	types.MsgProposal | types.MsgPrepare | types.MsgCommit | types.MsgRoundChange
}

type messagesByView[M msg] map[uint64]map[uint64][]*M

type feed struct {
	proposal    messagesByView[types.MsgProposal]
	prepare     messagesByView[types.MsgPrepare]
	commit      messagesByView[types.MsgCommit]
	roundChange messagesByView[types.MsgRoundChange]
}

type (
	singleRoundFeed feed
	validFeed       feed
)

func (f singleRoundFeed) SubscribeToProposalMessages(view *types.View, futureRounds bool) (<-chan func() []*types.MsgProposal, func()) {
	callback := func() []*types.MsgProposal {
		if futureRounds {
			return nil
		}

		return f.proposal[view.Sequence][view.Round]
	}

	c := make(chan func() []*types.MsgProposal, 1)
	c <- callback

	return c, func() { close(c) }
}

func (f singleRoundFeed) SubscribeToPrepareMessages(view *types.View, futureRounds bool) (<-chan func() []*types.MsgPrepare, func()) {
	callback := func() []*types.MsgPrepare {
		if futureRounds {
			return nil
		}

		return f.prepare[view.Sequence][view.Round]
	}

	c := make(chan func() []*types.MsgPrepare, 1)
	c <- callback

	return c, func() { close(c) }
}

func (f singleRoundFeed) SubscribeToCommitMessages(view *types.View, futureRounds bool) (<-chan func() []*types.MsgCommit, func()) {
	callback := func() []*types.MsgCommit {
		if futureRounds {
			return nil
		}

		return f.commit[view.Sequence][view.Round]
	}

	c := make(chan func() []*types.MsgCommit, 1)
	c <- callback

	return c, func() { close(c) }
}

func (f singleRoundFeed) SubscribeToRoundChangeMessages(view *types.View, futureRounds bool) (<-chan func() []*types.MsgRoundChange, func()) {
	callback := func() []*types.MsgRoundChange {
		if futureRounds {
			return nil
		}

		return f.roundChange[view.Sequence][view.Round]
	}

	c := make(chan func() []*types.MsgRoundChange, 1)
	c <- callback

	return c, func() { close(c) }
}

func (f validFeed) SubscribeToProposalMessages(view *types.View, higher bool) (<-chan func() []*types.MsgProposal, func()) {
	c := make(chan func() []*types.MsgProposal, 1)
	c <- func() []*types.MsgProposal {
		if higher == false {
			return f.proposal[view.Sequence][view.Round]
		}

		var max uint64
		for round := range f.proposal[view.Sequence] {
			if round >= max {
				max = round
			}
		}

		if max < view.Round {
			return nil
		}

		return f.proposal[view.Sequence][max]
	}

	return c, func() { close(c) }
}

func (f validFeed) SubscribeToPrepareMessages(view *types.View, higher bool) (<-chan func() []*types.MsgPrepare, func()) {
	c := make(chan func() []*types.MsgPrepare, 1)
	c <- func() []*types.MsgPrepare {
		if higher == false {
			return f.prepare[view.Sequence][view.Round]
		}

		var max uint64
		for round := range f.prepare[view.Sequence] {
			if round >= max {
				max = round
			}
		}

		if max < view.Round {
			return nil
		}

		return f.prepare[view.Sequence][max]
	}

	return c, func() { close(c) }
}

func (f validFeed) SubscribeToCommitMessages(view *types.View, higher bool) (<-chan func() []*types.MsgCommit, func()) {
	c := make(chan func() []*types.MsgCommit, 1)
	c <- func() []*types.MsgCommit {
		if higher == false {
			return f.commit[view.Sequence][view.Round]
		}

		var max uint64
		for round := range f.commit[view.Sequence] {
			if round >= max {
				max = round
			}
		}

		if max < view.Round {
			return nil
		}

		return f.commit[view.Sequence][max]
	}

	return c, func() { close(c) }
}

func (f validFeed) SubscribeToRoundChangeMessages(view *types.View, higher bool) (<-chan func() []*types.MsgRoundChange, func()) {
	c := make(chan func() []*types.MsgRoundChange, 1)
	c <- func() []*types.MsgRoundChange {
		if higher == false {
			return f.roundChange[view.Sequence][view.Round]
		}

		var max uint64
		for round := range f.roundChange[view.Sequence] {
			if round >= max {
				max = round
			}
		}

		if max < view.Round {
			return nil
		}

		return f.roundChange[view.Sequence][max]
	}

	return c, func() { close(c) }
}

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

func (v mockValidator) BuildBlock(uint64) []byte {
	return v.buildBlockFn()
}

type mockVerifier struct {
	isValidBlockFn func([]byte) bool
	isProposerFn   func([]byte, uint64, uint64) bool
	isValidatorFn  func([]byte, uint64) bool
}

func (v mockVerifier) IsValidBlock(block []byte, sequence uint64) bool {
	return v.isValidBlockFn(block)
}

func (v mockVerifier) IsProposer(id []byte, sequence uint64, round uint64) bool {
	return v.isProposerFn(id, sequence, round)
}

func (v mockVerifier) IsValidator(id []byte, height uint64) bool {
	return v.isValidatorFn(id, height)
}

type KeccakFn func([]byte) []byte

func (k KeccakFn) Hash(data []byte) []byte {
	return k(data)
}

type SigRecoverFn func([]byte, []byte) []byte

func (r SigRecoverFn) From(data, sig []byte) []byte {
	return r(data, sig)
}
