package sequencer

import (
	"context"

	"github.com/madz-lab/go-ibft"
)

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
