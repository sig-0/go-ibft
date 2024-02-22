package sequencer

import (
	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

type (
	QuorumFn     func(uint64, []ibft.Message) bool
	TransportFn  func(ibft.Message)
	KeccakFn     func([]byte) []byte
	SigRecoverFn func([]byte, []byte) []byte
)

func (q QuorumFn) HasQuorum(sequence uint64, msgs []ibft.Message) bool {
	return q(sequence, msgs)
}

func (t TransportFn) Multicast(msg ibft.Message) {
	t(msg)
}

func (k KeccakFn) Hash(data []byte) []byte {
	return k(data)
}

func (r SigRecoverFn) From(data, sig []byte) []byte {
	return r(data, sig)
}

type MockValidator struct {
	IDFn            func() []byte
	SignFn          func([]byte) []byte
	BuildProposalFn func() []byte
}

func (v MockValidator) ID() []byte {
	return v.IDFn()
}

func (v MockValidator) Sign(bytes []byte) []byte {
	return v.SignFn(bytes)
}

func (v MockValidator) BuildProposal(uint64) []byte {
	return v.BuildProposalFn()
}

type MockVerifier struct {
	HasValidSignatureFn func(ibft.Message) bool
	IsValidBlockFn      func([]byte) bool
	IsProposerFn        func([]byte, uint64, uint64) bool
	IsValidatorFn       func([]byte, uint64) bool
}

func (v MockVerifier) HasValidSignature(msg ibft.Message) bool {
	return v.HasValidSignatureFn(msg)
}

func (v MockVerifier) IsValidProposal(block []byte, _ uint64) bool {
	return v.IsValidBlockFn(block)
}

func (v MockVerifier) IsProposer(id []byte, sequence, round uint64) bool {
	return v.IsProposerFn(id, sequence, round)
}

func (v MockVerifier) IsValidator(id []byte, height uint64) bool {
	return v.IsValidatorFn(id, height)
}

type messagesByView[M types.IBFTMessage] map[uint64]map[uint64][]M

type feed struct {
	proposal    messagesByView[*types.MsgProposal]
	prepare     messagesByView[*types.MsgPrepare]
	commit      messagesByView[*types.MsgCommit]
	roundChange messagesByView[*types.MsgRoundChange]
}

type (
	singleRoundFeed feed // higher rounds disabled
	allRoundsFeed   feed // higher rounds enabled
)

func (f singleRoundFeed) ProposalMessages(view *types.View, futureRounds bool) (ibft.Subscription[*types.MsgProposal], func()) {
	callback := func() []*types.MsgProposal {
		if futureRounds {
			return nil
		}

		return f.proposal[view.Sequence][view.Round]
	}

	c := make(chan ibft.MsgNotification[*types.MsgProposal], 1)
	c <- ibft.NotificationFn[*types.MsgProposal](callback)

	return c, func() { close(c) }
}

func (f singleRoundFeed) PrepareMessages(view *types.View, futureRounds bool) (ibft.Subscription[*types.MsgPrepare], func()) {
	callback := func() []*types.MsgPrepare {
		if futureRounds {
			return nil
		}

		return f.prepare[view.Sequence][view.Round]
	}

	c := make(chan ibft.MsgNotification[*types.MsgPrepare], 1)
	c <- ibft.NotificationFn[*types.MsgPrepare](callback)

	return c, func() { close(c) }
}

func (f singleRoundFeed) CommitMessages(view *types.View, futureRounds bool) (ibft.Subscription[*types.MsgCommit], func()) {
	callback := func() []*types.MsgCommit {
		if futureRounds {
			return nil
		}

		return f.commit[view.Sequence][view.Round]
	}

	c := make(chan ibft.MsgNotification[*types.MsgCommit], 1)
	c <- ibft.NotificationFn[*types.MsgCommit](callback)

	return c, func() { close(c) }
}

func (f singleRoundFeed) RoundChangeMessages(view *types.View, futureRounds bool) (ibft.Subscription[*types.MsgRoundChange], func()) {
	callback := func() []*types.MsgRoundChange {
		if futureRounds {
			return nil
		}

		return f.roundChange[view.Sequence][view.Round]
	}

	c := make(chan ibft.MsgNotification[*types.MsgRoundChange], 1)
	c <- ibft.NotificationFn[*types.MsgRoundChange](callback)

	return c, func() { close(c) }
}

func (f allRoundsFeed) ProposalMessages(view *types.View, futureRounds bool) (ibft.Subscription[*types.MsgProposal], func()) {
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

	c := make(chan ibft.MsgNotification[*types.MsgProposal], 1)
	c <- ibft.NotificationFn[*types.MsgProposal](callback)

	return c, func() { close(c) }
}

func (f allRoundsFeed) PrepareMessages(view *types.View, futureRounds bool) (ibft.Subscription[*types.MsgPrepare], func()) {
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

	c := make(chan ibft.MsgNotification[*types.MsgPrepare], 1)
	c <- ibft.NotificationFn[*types.MsgPrepare](callback)

	return c, func() { close(c) }
}

func (f allRoundsFeed) CommitMessages(view *types.View, futureRounds bool) (ibft.Subscription[*types.MsgCommit], func()) {
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

	c := make(chan ibft.MsgNotification[*types.MsgCommit], 1)
	c <- ibft.NotificationFn[*types.MsgCommit](callback)

	return c, func() { close(c) }
}

func (f allRoundsFeed) RoundChangeMessages(view *types.View, futureRounds bool) (ibft.Subscription[*types.MsgRoundChange], func()) {
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

	c := make(chan ibft.MsgNotification[*types.MsgRoundChange], 1)
	c <- ibft.NotificationFn[*types.MsgRoundChange](callback)

	return c, func() { close(c) }
}
