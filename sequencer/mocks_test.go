package sequencer

import (
	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

type (
	QuorumFn     func([]ibft.Message) bool
	TransportFn  func(ibft.Message)
	KeccakFn     func([]byte) []byte
	SigRecoverFn func([]byte, []byte) []byte

	MockValidator struct {
		IDFn            func() []byte
		SignFn          func([]byte) []byte
		BuildProposalFn func() []byte
	}

	MockVerifier struct {
		IsValidSignatureFn func([]byte, []byte, []byte) bool
		IsValidBlockFn     func([]byte) bool
		IsProposerFn       func([]byte, uint64, uint64) bool
		IsValidatorFn      func([]byte, uint64) bool
	}

	messagesByView[M types.IBFTMessage] map[uint64]map[uint64][]M

	MockFeed struct {
		proposal    messagesByView[*types.MsgProposal]
		prepare     messagesByView[*types.MsgPrepare]
		commit      messagesByView[*types.MsgCommit]
		roundChange messagesByView[*types.MsgRoundChange]
	}

	singleRoundFeed MockFeed // higher rounds disabled
	allRoundsFeed   MockFeed // higher rounds enabled
)

func (q QuorumFn) HasQuorum(messages []ibft.Message) bool {
	return q(messages)
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

func (v MockValidator) ID() []byte {
	return v.IDFn()
}

func (v MockValidator) Sign(bytes []byte) []byte {
	return v.SignFn(bytes)
}

func (v MockValidator) BuildProposal(uint64) []byte {
	return v.BuildProposalFn()
}

func (v MockVerifier) IsValidSignature(sender []byte, digest []byte, sig []byte) bool {
	return v.IsValidSignatureFn(sender, digest, sig)
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

func newSingleRoundSubscription[M types.IBFTMessage](
	messages messagesByView[M],
	view *types.View,
	futureRounds bool,
) (ibft.Subscription[M], func()) {
	c := make(chan ibft.MsgNotification[M], 1)
	c <- ibft.NotificationFn[M](func() []M {
		if futureRounds {
			return nil
		}

		return messages[view.Sequence][view.Round]
	})

	return c, func() { close(c) }
}

func newFutureRoundsSubscription[M types.IBFTMessage](
	messages messagesByView[M],
	view *types.View,
	futureRounds bool,
) (ibft.Subscription[M], func()) {
	c := make(chan ibft.MsgNotification[M], 1)
	c <- ibft.NotificationFn[M](func() []M {
		if futureRounds == false {
			return messages[view.Sequence][view.Round]
		}

		var max uint64
		for round := range messages[view.Sequence] {
			if round >= max {
				max = round
			}
		}

		if max < view.Round {
			return nil
		}

		return messages[view.Sequence][max]
	})

	return c, func() { close(c) }
}

func (f singleRoundFeed) ProposalMessages(
	view *types.View,
	futureRounds bool,
) (ibft.Subscription[*types.MsgProposal], func()) {
	return newSingleRoundSubscription[*types.MsgProposal](f.proposal, view, futureRounds)
}

func (f singleRoundFeed) PrepareMessages(
	view *types.View,
	futureRounds bool,
) (ibft.Subscription[*types.MsgPrepare], func()) {
	return newSingleRoundSubscription[*types.MsgPrepare](f.prepare, view, futureRounds)
}

func (f singleRoundFeed) CommitMessages(
	view *types.View,
	futureRounds bool,
) (ibft.Subscription[*types.MsgCommit], func()) {
	return newSingleRoundSubscription[*types.MsgCommit](f.commit, view, futureRounds)
}

func (f singleRoundFeed) RoundChangeMessages(
	view *types.View,
	futureRounds bool,
) (ibft.Subscription[*types.MsgRoundChange], func()) {
	return newSingleRoundSubscription[*types.MsgRoundChange](f.roundChange, view, futureRounds)
}

func (f allRoundsFeed) ProposalMessages(
	view *types.View,
	futureRounds bool,
) (ibft.Subscription[*types.MsgProposal], func()) {
	return newFutureRoundsSubscription[*types.MsgProposal](f.proposal, view, futureRounds)
}

func (f allRoundsFeed) PrepareMessages(
	view *types.View,
	futureRounds bool,
) (ibft.Subscription[*types.MsgPrepare], func()) {
	return newFutureRoundsSubscription[*types.MsgPrepare](f.prepare, view, futureRounds)
}

func (f allRoundsFeed) CommitMessages(
	view *types.View,
	futureRounds bool,
) (ibft.Subscription[*types.MsgCommit], func()) {
	return newFutureRoundsSubscription[*types.MsgCommit](f.commit, view, futureRounds)
}

func (f allRoundsFeed) RoundChangeMessages(
	view *types.View,
	futureRounds bool,
) (ibft.Subscription[*types.MsgRoundChange], func()) {
	return newFutureRoundsSubscription[*types.MsgRoundChange](f.roundChange, view, futureRounds)
}
