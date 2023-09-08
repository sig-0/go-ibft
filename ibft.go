//nolint:forcetypeassert, gocritic
package ibft

import (
	"context"

	"github.com/madz-lab/go-ibft/message/types"
)

/** External functionalities required by the IBFT 2.0 block finalization algorithm	**/

type (
	// Signer generates a signature based on some arbitrary input
	Signer = types.Signer

	// SigRecover is used to extract the address associated with some payload and its signature
	SigRecover = types.SigRecover

	// Feed provides streams to listen for consensus messages gossiped by the network
	Feed = types.MsgFeed

	// Transport is used to gossip a consensus message to the network
	Transport interface {
		// Multicast gossips a consensus message to the network
		Multicast(types.Msg)
	}

	// Quorum asserts if the provided messages indicate a reach of consensus in a particular sequence
	Quorum interface {
		// HasQuorum checks if quorum has been reached for given sequence
		HasQuorum(uint64, []types.Msg) bool
	}

	// Keccak is used to obtain the Keccak encoding of arbitrary data
	Keccak interface {
		// Hash returns the Keccak encoding of given input
		Hash([]byte) []byte
	}

	// Verifier authenticates message senders and block data
	Verifier interface {
		// IsValidator checks if id is part of consensus for given sequence
		IsValidator(id []byte, sequence uint64) bool

		// IsValidBlock checks if the provided block is valid for given sequence
		IsValidBlock(block []byte, sequence uint64) bool

		// IsProposer checks if the id is the select proposer for given sequence and round
		IsProposer(id []byte, sequence uint64, round uint64) bool
	}

	// Validator is one of the uniquely identified consensus actors in the network.
	// It partakes in block finalization by signing consensus message and proposing blocks
	Validator interface {
		Signer

		// ID returns the unique ID of this validator
		ID() []byte

		// BuildBlock returns a byte representation of a potential block for given sequence
		BuildBlock(uint64) []byte
	}
)

type ctxKey string

const (
	// Context keys
	transport  ctxKey = "transport"
	feed       ctxKey = "feed"
	quorum     ctxKey = "quorum"
	keccak     ctxKey = "keccak"
	sigRecover ctxKey = "sig_recover"
)

// Context is a convenience wrapper that provides external functionalities
// to the finalization algorithm (sequencer). This context is only meant to be cancelled by
// by the user and is never cancelled by the algorithm itself
type Context struct {
	context.Context
}

// NewIBFTContext wraps the context.Context with IBFT specific methods
func NewIBFTContext(ctx context.Context) Context {
	return Context{ctx}
}

// WithCancel returns a wrapped child context
func (c Context) WithCancel() (Context, func()) {
	subCtx, cancelFn := context.WithCancel(c)

	return Context{subCtx}, cancelFn
}

// WithTransport sets the required gossip functionality
func (c Context) WithTransport(t Transport) Context {
	return Context{context.WithValue(c, transport, t)}
}

// Transport returns the Transport instance associated with this context
func (c Context) Transport() Transport {
	return c.Value(transport).(Transport)
}

// WithFeed sets the required stream for incoming messages
func (c Context) WithFeed(f Feed) Context {
	return Context{context.WithValue(c, feed, f)}
}

// Feed returns the Feed instance associated with this context
func (c Context) Feed() Feed {
	return c.Value(feed).(Feed)
}

// WithQuorum sets the required check for reach of consensus
func (c Context) WithQuorum(q Quorum) Context {
	return Context{context.WithValue(c, quorum, q)}
}

// Quorum returns the consensus checker associated with this context
func (c Context) Quorum() Quorum {
	return c.Value(quorum).(Quorum)
}

// WithKeccak sets the required hash generator
func (c Context) WithKeccak(k Keccak) Context {
	return Context{context.WithValue(c, keccak, k)}
}

// Keccak returns the hash generator associated with this context
func (c Context) Keccak() Keccak {
	return c.Value(keccak).(Keccak)
}

// WithSigRecover sets the required address recovery mechanism
func (c Context) WithSigRecover(s SigRecover) Context {
	return Context{context.WithValue(c, sigRecover, s)}
}

// SigRecover returns the address recovery mechanism associated with this context
func (c Context) SigRecover() SigRecover {
	return c.Value(sigRecover).(SigRecover)
}
