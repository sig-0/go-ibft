//nolint:all
package sequencer

import (
	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/store"
	"github.com/madz-lab/go-ibft/message/types"
)

type QuorumFn func(uint64, []ibft.Message) bool

func (q QuorumFn) HasQuorum(sequence uint64, msgs []ibft.Message) bool {
	return q(sequence, msgs)
}

type TransportFn func(ibft.Message)

func (t TransportFn) Multicast(msg ibft.Message) {
	t(msg)
}

type KeccakFn func([]byte) []byte

func (k KeccakFn) Hash(data []byte) []byte {
	return k(data)
}

type SigRecoverFn func([]byte, []byte) []byte

func (r SigRecoverFn) From(data, sig []byte) []byte {
	return r(data, sig)
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

func (v mockValidator) BuildProposal(uint64) []byte {
	return v.buildBlockFn()
}

type mockVerifier struct {
	hasValidSignatureFn func(ibft.Message) bool
	isValidBlockFn      func([]byte) bool
	isProposerFn        func([]byte, uint64, uint64) bool
	isValidatorFn       func([]byte, uint64) bool
}

func (v mockVerifier) HasValidSignature(msg ibft.Message) bool {
	return v.hasValidSignatureFn(msg)
}

func (v mockVerifier) IsValidProposal(block []byte, _ uint64) bool {
	return v.isValidBlockFn(block)
}

func (v mockVerifier) IsProposer(id []byte, sequence, round uint64) bool {
	return v.isProposerFn(id, sequence, round)
}

func (v mockVerifier) IsValidator(id []byte, height uint64) bool {
	return v.isValidatorFn(id, height)
}

type messagesByView[M types.IBFTMessage] map[uint64]map[uint64][]*M

type feed struct {
	proposal    messagesByView[types.MsgProposal]
	prepare     messagesByView[types.MsgPrepare]
	commit      messagesByView[types.MsgCommit]
	roundChange messagesByView[types.MsgRoundChange]
}

type (
	singleRoundFeed feed // higher rounds disabled
	allRoundsFeed   feed // higher rounds enabled
)

func (f singleRoundFeed) ProposalMessages(view *types.View, futureRounds bool) (store.Subscription[types.MsgProposal], func()) {
	callback := func() []*types.MsgProposal {
		if futureRounds {
			return nil
		}

		return f.proposal[view.Sequence][view.Round]
	}

	c := make(chan store.MsgNotification[types.MsgProposal], 1)
	c <- store.MsgReceiverFn[types.MsgProposal](callback)

	return c, func() { close(c) }
}

func (f singleRoundFeed) PrepareMessages(view *types.View, futureRounds bool) (store.Subscription[types.MsgPrepare], func()) {
	callback := func() []*types.MsgPrepare {
		if futureRounds {
			return nil
		}

		return f.prepare[view.Sequence][view.Round]
	}

	c := make(chan store.MsgNotification[types.MsgPrepare], 1)
	c <- store.MsgReceiverFn[types.MsgPrepare](callback)

	return c, func() { close(c) }
}

func (f singleRoundFeed) CommitMessages(view *types.View, futureRounds bool) (store.Subscription[types.MsgCommit], func()) {
	callback := func() []*types.MsgCommit {
		if futureRounds {
			return nil
		}

		return f.commit[view.Sequence][view.Round]
	}

	c := make(chan store.MsgNotification[types.MsgCommit], 1)
	c <- store.MsgReceiverFn[types.MsgCommit](callback)

	return c, func() { close(c) }
}

func (f singleRoundFeed) RoundChangeMessages(view *types.View, futureRounds bool) (store.Subscription[types.MsgRoundChange], func()) {
	callback := func() []*types.MsgRoundChange {
		if futureRounds {
			return nil
		}

		return f.roundChange[view.Sequence][view.Round]
	}

	c := make(chan store.MsgNotification[types.MsgRoundChange], 1)
	c <- store.MsgReceiverFn[types.MsgRoundChange](callback)

	return c, func() { close(c) }
}

func (f allRoundsFeed) ProposalMessages(view *types.View, futureRounds bool) (store.Subscription[types.MsgProposal], func()) {
	callback := func() []*types.MsgProposal {
		if futureRounds == false {
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

	c := make(chan store.MsgNotification[types.MsgProposal], 1)
	c <- store.MsgReceiverFn[types.MsgProposal](callback)

	return c, func() { close(c) }
}

func (f allRoundsFeed) PrepareMessages(view *types.View, futureRounds bool) (store.Subscription[types.MsgPrepare], func()) {
	callback := func() []*types.MsgPrepare {
		if futureRounds == false {
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

	c := make(chan store.MsgNotification[types.MsgPrepare], 1)
	c <- store.MsgReceiverFn[types.MsgPrepare](callback)

	return c, func() { close(c) }
}

func (f allRoundsFeed) CommitMessages(view *types.View, futureRounds bool) (store.Subscription[types.MsgCommit], func()) {
	callback := func() []*types.MsgCommit {
		if futureRounds == false {
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

	c := make(chan store.MsgNotification[types.MsgCommit], 1)
	c <- store.MsgReceiverFn[types.MsgCommit](callback)

	return c, func() { close(c) }
}

func (f allRoundsFeed) RoundChangeMessages(view *types.View, futureRounds bool) (store.Subscription[types.MsgRoundChange], func()) {
	callback := func() []*types.MsgRoundChange {
		if futureRounds == false {
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

	c := make(chan store.MsgNotification[types.MsgRoundChange], 1)
	c <- store.MsgReceiverFn[types.MsgRoundChange](callback)

	return c, func() { close(c) }
}
