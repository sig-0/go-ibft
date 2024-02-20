package store

import (
	"github.com/madz-lab/go-ibft/message/types"
)

type Subscription[M types.IBFTMessage] chan MsgNotification[M]

type MsgNotification[M types.IBFTMessage] interface {
	Receive() []*M
}

type MsgReceiverFn[M types.IBFTMessage] func() []*M

func (r MsgReceiverFn[M]) Receive() []*M {
	return r()
}

// MsgFeed provides an asynchronous way to receive consensus messages. In addition
// to listen for any type of message for any particular view, the higherRounds flag provides an option
// to receive messages from rounds higher than the round in provided view.
//
// CONTRACT: messages received by consuming the channel's callback are assumed to be valid:
//
// - any message has a valid view (matches the one provided)
//
// - all messages are considered unique (there cannot be 2 or more messages with identical From fields)
type MsgFeed interface {
	// ProposalMessages returns the MsgProposal subscription for given view(s)
	ProposalMessages(view *types.View, higherRounds bool) (Subscription[types.MsgProposal], func())

	// PrepareMessages returns the MsgPrepare subscription for given view(s)
	PrepareMessages(view *types.View, higherRounds bool) (Subscription[types.MsgPrepare], func())

	// CommitMessages returns the MsgCommit subscription for given view(s)
	CommitMessages(view *types.View, higherRounds bool) (Subscription[types.MsgCommit], func())

	// RoundChangeMessages returns the MsgRoundChange subscription for given view(s)
	RoundChangeMessages(view *types.View, higherRounds bool) (Subscription[types.MsgRoundChange], func())
}

type feed struct {
	*Store
}

func (f feed) ProposalMessages(view *types.View, futureRounds bool) (Subscription[types.MsgProposal], func()) {
	return f.Store.proposal.Subscribe(view, futureRounds)
}

func (f feed) PrepareMessages(view *types.View, futureRounds bool) (Subscription[types.MsgPrepare], func()) {
	return f.Store.prepare.Subscribe(view, futureRounds)
}

func (f feed) CommitMessages(view *types.View, futureRounds bool) (Subscription[types.MsgCommit], func()) {
	return f.Store.commit.Subscribe(view, futureRounds)
}

func (f feed) RoundChangeMessages(view *types.View, futureRounds bool) (Subscription[types.MsgRoundChange], func()) {
	return f.Store.roundChange.Subscribe(view, futureRounds)
}
