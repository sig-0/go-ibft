package sequencer

import (
	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

func newSingleRoundFeed(messages []ibft.Message) MessageFeed {
	feed := singleRoundFeed{
		proposal:    make(messagesByView[*types.MsgProposal]),
		prepare:     make(messagesByView[*types.MsgPrepare]),
		commit:      make(messagesByView[*types.MsgCommit]),
		roundChange: make(messagesByView[*types.MsgRoundChange]),
	}

	for _, msg := range messages {
		switch msg := msg.(type) {
		case *types.MsgProposal:
			feed.proposal.add(msg)
		case *types.MsgPrepare:
			feed.prepare.add(msg)
		case *types.MsgCommit:
			feed.commit.add(msg)
		case *types.MsgRoundChange:
			feed.roundChange.add(msg)
		}
	}

	return feed
}

func newAllRoundsFeed(messages []ibft.Message) MessageFeed {
	feed := allRoundsFeed{
		proposal:    make(messagesByView[*types.MsgProposal]),
		prepare:     make(messagesByView[*types.MsgPrepare]),
		commit:      make(messagesByView[*types.MsgCommit]),
		roundChange: make(messagesByView[*types.MsgRoundChange]),
	}

	for _, msg := range messages {
		switch msg := msg.(type) {
		case *types.MsgProposal:
			feed.proposal.add(msg)
		case *types.MsgPrepare:
			feed.prepare.add(msg)
		case *types.MsgCommit:
			feed.commit.add(msg)
		case *types.MsgRoundChange:
			feed.roundChange.add(msg)
		}
	}

	return feed
}

type (
	messagesByView[M msg] map[uint64]map[uint64][]M

	MockFeed struct {
		proposal    messagesByView[*types.MsgProposal]
		prepare     messagesByView[*types.MsgPrepare]
		commit      messagesByView[*types.MsgCommit]
		roundChange messagesByView[*types.MsgRoundChange]
	}

	singleRoundFeed MockFeed // higher rounds disabled
	allRoundsFeed   MockFeed // higher rounds enabled
)

type msg interface {
	types.IBFTMessage

	GetView() *types.View
}

func (m messagesByView[M]) add(msg M) {
	view := msg.GetView()

	messagesInSequence, ok := m[view.Sequence]
	if !ok {
		m[view.Sequence] = make(map[uint64][]M)
		messagesInSequence = m[view.Sequence]
	}

	messagesInRound, ok := messagesInSequence[view.Round]
	if !ok {
		messagesInSequence[view.Round] = make([]M, 0)
		messagesInRound = messagesInSequence[view.Round]
	}

	messagesInRound = append(messagesInRound, msg)
	messagesInSequence[view.Round] = messagesInRound
}

func newSingleRoundSubscription[M msg](
	messages messagesByView[M],
	view *types.View,
	futureRounds bool,
) (types.Subscription[M], func()) {
	c := make(chan types.MsgNotification[M], 1)
	c <- types.NotificationFn[M](func() []M {
		if futureRounds {
			return nil
		}

		return messages[view.Sequence][view.Round]
	})

	return c, func() { close(c) }
}

func newFutureRoundsSubscription[M msg](
	messages messagesByView[M],
	view *types.View,
	futureRounds bool,
) (types.Subscription[M], func()) {
	c := make(chan types.MsgNotification[M], 1)
	c <- types.NotificationFn[M](func() []M {
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
) (types.Subscription[*types.MsgProposal], func()) {
	return newSingleRoundSubscription[*types.MsgProposal](f.proposal, view, futureRounds)
}

func (f singleRoundFeed) PrepareMessages(
	view *types.View,
	futureRounds bool,
) (types.Subscription[*types.MsgPrepare], func()) {
	return newSingleRoundSubscription[*types.MsgPrepare](f.prepare, view, futureRounds)
}

func (f singleRoundFeed) CommitMessages(
	view *types.View,
	futureRounds bool,
) (types.Subscription[*types.MsgCommit], func()) {
	return newSingleRoundSubscription[*types.MsgCommit](f.commit, view, futureRounds)
}

func (f singleRoundFeed) RoundChangeMessages(
	view *types.View,
	futureRounds bool,
) (types.Subscription[*types.MsgRoundChange], func()) {
	return newSingleRoundSubscription[*types.MsgRoundChange](f.roundChange, view, futureRounds)
}

func (f allRoundsFeed) ProposalMessages(
	view *types.View,
	futureRounds bool,
) (types.Subscription[*types.MsgProposal], func()) {
	return newFutureRoundsSubscription[*types.MsgProposal](f.proposal, view, futureRounds)
}

func (f allRoundsFeed) PrepareMessages(
	view *types.View,
	futureRounds bool,
) (types.Subscription[*types.MsgPrepare], func()) {
	return newFutureRoundsSubscription[*types.MsgPrepare](f.prepare, view, futureRounds)
}

func (f allRoundsFeed) CommitMessages(
	view *types.View,
	futureRounds bool,
) (types.Subscription[*types.MsgCommit], func()) {
	return newFutureRoundsSubscription[*types.MsgCommit](f.commit, view, futureRounds)
}

func (f allRoundsFeed) RoundChangeMessages(
	view *types.View,
	futureRounds bool,
) (types.Subscription[*types.MsgRoundChange], func()) {
	return newFutureRoundsSubscription[*types.MsgRoundChange](f.roundChange, view, futureRounds)
}
