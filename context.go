package ibft

import "context"

type ctxKey string

const (
	transport  ctxKey = "transport"
	feed       ctxKey = "feed"
	quorum     ctxKey = "quorum"
	keccak     ctxKey = "keccak"
	sigRecover ctxKey = "sig_recover"
)

// Context is a convenience wrapper that provides external functionalities
// to the finalization algorithm (sequencer.Sequencer). This context is only meant to be cancelled by
// by the user and is never cancelled by the protocol itself
type Context struct {
	context.Context
}

// NewIBFTContext wraps the provided context with IBFT api
func NewIBFTContext(ctx context.Context) Context {
	return Context{ctx}
}

// WithCancel returns a wrapped child context with cancellation func
func (c Context) WithCancel() (Context, func()) {
	subCtx, cancelFn := context.WithCancel(c)

	return Context{subCtx}, cancelFn
}

// WithTransport sets the required message transport functionality
func (c Context) WithTransport(t Transport) Context {
	return Context{context.WithValue(c, transport, t)}
}

// Transport returns the Transport instance associated with this context
func (c Context) Transport() Transport {
	return c.Value(transport).(Transport) //nolint:forcetypeassert // redundant
}

// WithFeed sets the required message feed
func (c Context) WithFeed(f MessageFeed) Context {
	return Context{context.WithValue(c, feed, f)}
}

// Feed returns the Feed instance associated with this context
func (c Context) Feed() MessageFeed {
	return c.Value(feed).(MessageFeed) //nolint:forcetypeassert // redundant
}

// WithQuorum sets the required Quorum callback to check for a reach in consensus
func (c Context) WithQuorum(q Quorum) Context {
	return Context{context.WithValue(c, quorum, q)}
}

// Quorum returns the Quorum callback associated with this context
func (c Context) Quorum() Quorum {
	return c.Value(quorum).(Quorum) //nolint:forcetypeassert // redundant
}

// WithKeccak sets the required Keccak hash generator
func (c Context) WithKeccak(k Keccak) Context {
	return Context{context.WithValue(c, keccak, k)}
}

// Keccak returns the Keccak hash generator associated with this context
func (c Context) Keccak() Keccak {
	return c.Value(keccak).(Keccak) //nolint:forcetypeassert // redundant
}

// WithSigRecover sets the required sender recovery callback
func (c Context) WithSigRecover(s SigRecover) Context {
	return Context{context.WithValue(c, sigRecover, s)}
}

// SigRecover returns the sender recovery callback associated with this context
func (c Context) SigRecover() SigRecover {
	return c.Value(sigRecover).(SigRecover) //nolint:forcetypeassert // redundant
}
