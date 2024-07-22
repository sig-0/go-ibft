package sequencer

import (
	"context"

	"github.com/sig-0/go-ibft"
)

type ctxKey string

const (
	transport ctxKey = "transport"
	feed      ctxKey = "feed"
	quorum    ctxKey = "quorum"
	keccak    ctxKey = "keccak"
)

// Context is a convenience wrapper that provides external functionalities
// to the finalization algorithm run by Sequencer. Context is never cancelled by the protocol, only by the caller.
type Context struct {
	context.Context
}

func NewContext(ctx context.Context) Context {
	return Context{ctx}
}

func (c Context) WithKeccak(k ibft.Keccak) Context {
	return Context{context.WithValue(c, keccak, k)}
}

func (c Context) WithQuorum(q ibft.Quorum) Context {
	return Context{context.WithValue(c, quorum, q)}
}

func (c Context) WithTransport(t ibft.MsgTransport) Context {
	return Context{context.WithValue(c, transport, t)}
}

func (c Context) WithMsgFeed(f MsgFeed) Context {
	return Context{context.WithValue(c, feed, f)}
}

func (c Context) Keccak() ibft.Keccak {
	return c.Value(keccak).(ibft.Keccak) //nolint:forcetypeassert // already wrapped
}

func (c Context) Quorum() ibft.Quorum {
	return c.Value(quorum).(ibft.Quorum) //nolint:forcetypeassert // already wrapped
}

func (c Context) Transport() ibft.MsgTransport {
	return c.Value(transport).(ibft.MsgTransport) //nolint:forcetypeassert // redundant
}

func (c Context) MessageFeed() MsgFeed {
	return c.Value(feed).(MsgFeed) //nolint:forcetypeassert // redundant
}
