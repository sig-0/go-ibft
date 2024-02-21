package ibft

import (
	"context"

	"github.com/madz-lab/go-ibft/message/store"
	"github.com/madz-lab/go-ibft/message/types"
)

// Message is a convenience wrapper for the IBFT consensus messages. See types.IBFTMessage for concrete types
type Message interface {
	// GetFrom returns the address associated with this Message
	GetFrom() []byte

	// GetSignature returns the signature of this Message
	GetSignature() []byte

	// Payload returns the byte content of this Message (signature excluded)
	Payload() []byte

	// Bytes returns the byte content of this Message (signature included)
	Bytes() []byte
}

/** External functionalities required by the IBFT block finalization algorithm	**/

type (
	// Signer is used to generate a sender-unique signature of arbitrary data
	Signer interface {
		// Sign computes the sender-unique signature of data
		Sign(data []byte) []byte
	}

	// SigRecover is used to extract the sender associated with some payload and its signature
	SigRecover interface {
		// From returns the sender associated with this data and signature
		From(data, sig []byte) []byte
	}

	// Feed provides streaming objects for consensus messages gossiped by the network
	Feed = store.MsgFeed

	// Transport is used to gossip a consensus message to the network
	Transport interface {
		// Multicast gossips a consensus message to the network
		Multicast(Message)
	}

	// Quorum is used to check whether the protocol reached consensus during a particular step
	Quorum interface {
		// HasQuorum returns true if messages accumulate consensus for a particular sequence
		HasQuorum(sequence uint64, messages []Message) bool
	}

	// Keccak is used to obtain the Keccak encoding of arbitrary data
	Keccak interface {
		// Hash returns the Keccak encoding of given input
		Hash([]byte) []byte
	}

	// Verifier authenticates message senders and block data
	Verifier interface {
		// HasValidSignature checks if the signature of the message is valid
		HasValidSignature(msg Message) bool

		// IsValidator checks if id is part of consensus for given sequence
		IsValidator(id []byte, sequence uint64) bool

		// IsValidProposal checks if the provided proposal is valid for given sequence
		IsValidProposal(proposal []byte, sequence uint64) bool

		// IsProposer checks if the id is the elected proposer for given sequence and round
		IsProposer(id []byte, sequence, round uint64) bool
	}

	// Validator represents one uniquely identified consensus actor in the network.
	// Its role in the protocol is to broadcast (signed) consensus messages and make proposals (if elected)
	Validator interface {
		Signer

		// ID returns the unique ID of this validator
		ID() []byte

		// BuildProposal returns this validator's proposal for given sequence
		BuildProposal(sequence uint64) []byte
	}
)

// WrapMessages wraps concrete message types into Message type
func WrapMessages[M types.IBFTMessage](messages []M) []Message {
	wrapped := make([]Message, 0, len(messages))

	for _, msg := range messages {
		//nolint:errcheck // guarded by the types.IBFTMessage constraint
		wrapped = append(wrapped, any(msg).(Message))
	}

	return wrapped
}

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
func (c Context) WithFeed(f Feed) Context {
	return Context{context.WithValue(c, feed, f)}
}

// Feed returns the Feed instance associated with this context
func (c Context) Feed() Feed {
	return c.Value(feed).(Feed) //nolint:forcetypeassert // redundant
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
