package sequencer

import (
	"context"
	"github.com/madz-lab/go-ibft"
)

// Context is a convenience wrapper that provides external functionalities
// to the finalization algorithm (sequencer.Sequencer). This context is only meant to be cancelled by
// by the user and is never cancelled by the protocol itself
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

func (c Context) Transport() ibft.Transport {
	return c.Value(transport).(ibft.Transport)
}

func (c Context) Quorum() ibft.Quorum {
	return c.Value(quorum).(ibft.Quorum)
}

func (c Context) Feed() ibft.MessageFeed {
	return c.Value(feed).(ibft.MessageFeed)
}

type ctxKey string

const (
	transport ctxKey = "transport"
	feed      ctxKey = "feed"
	quorum    ctxKey = "quorum"
	keccak    ctxKey = "keccak"
)

type ContextOption func(Context) Context

func WithTransport(t ibft.Transport) ContextOption {
	return func(c Context) Context {
		return Context{context.WithValue(c, transport, t)}
	}
}

func WithMessageFeed(f ibft.MessageFeed) ContextOption {
	return func(c Context) Context {
		return Context{context.WithValue(c, feed, f)}
	}
}

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
