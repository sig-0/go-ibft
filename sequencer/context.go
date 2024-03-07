package sequencer

import (
	"context"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

// MessageFeed provides an asynchronous way to receive consensus messages. In addition
// to listen for any type of message for any particular view, the higherRounds flag provides an option
// to receive messages from rounds higher than the round in provided view.
//
// CONTRACT: messages received by consuming the channel's callback are assumed to be valid:
//
// - any message has a valid view (matches the one provided)
//
// - all messages are considered unique (there cannot be 2 or more messages with identical From fields)
type MessageFeed interface {
	// ProposalMessages returns the MsgProposal subscription for given view(s)
	ProposalMessages(view *types.View, higherRounds bool) (types.Subscription[*types.MsgProposal], func())

	// PrepareMessages returns the MsgPrepare subscription for given view(s)
	PrepareMessages(view *types.View, higherRounds bool) (types.Subscription[*types.MsgPrepare], func())

	// CommitMessages returns the MsgCommit subscription for given view(s)
	CommitMessages(view *types.View, higherRounds bool) (types.Subscription[*types.MsgCommit], func())

	// RoundChangeMessages returns the MsgRoundChange subscription for given view(s)
	RoundChangeMessages(view *types.View, higherRounds bool) (types.Subscription[*types.MsgRoundChange], func())
}

type MessageTransport struct {
	Proposal    ibft.Transport[*types.MsgProposal]
	Prepare     ibft.Transport[*types.MsgPrepare]
	Commit      ibft.Transport[*types.MsgCommit]
	RoundChange ibft.Transport[*types.MsgRoundChange]
}

type ctxKey string

const (
	transport ctxKey = "transport"
	feed      ctxKey = "feed"
	quorum    ctxKey = "quorum"
	keccak    ctxKey = "keccak"
)

type ContextOption func(Context) Context

func WithQuorum(q ibft.Quorum) ContextOption {
	return func(c Context) Context {
		return Context{context.WithValue(c, quorum, q)}
	}
}

func WithKeccak(k ibft.Keccak) ContextOption {
	return func(c Context) Context {
		return Context{context.WithValue(c, keccak, k)}
	}
}

func WithMessageFeed(f MessageFeed) ContextOption {
	return func(c Context) Context {
		return Context{context.WithValue(c, feed, f)}
	}
}

func WithMessageTransport(t MessageTransport) ContextOption {
	return func(c Context) Context {
		return Context{context.WithValue(c, transport, t)}
	}
}

// Context is a convenience wrapper that provides external functionalities
// to the finalization algorithm (Sequencer). This Context is meant to be cancelled
// only by the caller and is never cancelled by the protocol itself
type Context struct {
	context.Context
}

func NewContext(ctx context.Context, opts ...ContextOption) Context {
	c := Context{ctx}

	for _, opt := range opts {
		c = opt(c)
	}

	return c
}

func (c Context) Keccak() ibft.Keccak {
	return c.Value(keccak).(ibft.Keccak)
}

func (c Context) Quorum() ibft.Quorum {
	return c.Value(quorum).(ibft.Quorum)
}

func (c Context) MessageFeed() MessageFeed {
	return c.Value(feed).(MessageFeed)
}

func (c Context) MessageTransport() MessageTransport {
	return c.Value(transport).(MessageTransport)
}
