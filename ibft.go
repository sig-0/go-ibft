package ibft

import "github.com/madz-lab/go-ibft/message/types"

// Message is a convenience wrapper for the IBFT consensus messages. See types.IBFTMessage for concrete types
type Message interface {
	// GetSequence returns the sequence for which this Message was created
	GetSequence() uint64

	// GetRound returns the round in which this Message was created
	GetRound() uint64

	// GetSender returns the sender ID associated with this Message
	GetSender() []byte

	// GetSignature returns the signature of this Message
	GetSignature() []byte

	// Payload returns the byte content of this Message (signature excluded)
	Payload() []byte

	// Bytes returns the byte content of this Message (signature included)
	Bytes() []byte
}

// Validator represents a unique consensus actor in the network.
// Its role in the protocol is to broadcast (signed) consensus messages and make proposals (if elected)
type Validator interface {
	Signer
	Verifier

	// ID returns the network-unique ID of this validator
	ID() []byte

	// BuildProposal returns this validator's proposal for given sequence
	BuildProposal(sequence uint64) []byte
}

/** External functionalities required by the IBFT block finalization algorithm	**/

type (
	// Verifier authenticates consensus messages gossiped in the network
	Verifier interface {
		// IsValidSignature checks if the signature of the message is valid
		IsValidSignature(sender, digest, sig []byte) bool

		// IsValidator checks if id is part of consensus for given sequence
		IsValidator(id []byte, sequence uint64) bool

		// IsValidProposal checks if the provided proposal is valid for given sequence
		IsValidProposal(proposal []byte, sequence uint64) bool

		// IsProposer asserts if id is the elected proposer for given sequence and round
		IsProposer(id []byte, sequence, round uint64) bool
	}

	// Signer is used to generate a sender-unique signature of a keccak hash
	Signer interface {
		// Sign computes the signature of given digest
		Sign(digest []byte) []byte
	}

	SignerFn func([]byte) []byte

	// Transport is used to gossip a consensus message to the network
	Transport[M types.IBFTMessage] interface {
		// Multicast gossips a consensus message to the network
		Multicast(M)
	}

	TransportFn[M types.IBFTMessage] func(msg M)

	// Quorum is used to check whether the protocol reached consensus during a particular step
	Quorum interface {
		// HasQuorum returns true if messages accumulate consensus for a particular sequence
		HasQuorum(messages []Message) bool
	}

	QuorumFn func([]Message) bool

	// Keccak is used to obtain the Keccak encoding of arbitrary data
	Keccak interface {
		// Hash returns the Keccak encoding of given input
		Hash(payload []byte) []byte
	}

	KeccakFn func([]byte) []byte
)

func (f TransportFn[M]) Multicast(msg M) {
	f(msg)
}

func (f KeccakFn) Hash(digest []byte) []byte {
	return f(digest)
}

func (f SignerFn) Sign(digest []byte) []byte {
	return f(digest)
}

func (f QuorumFn) HasQuorum(messages []Message) bool {
	return f(messages)
}

// WrapMessages wraps concrete message types into Message type
func WrapMessages[M types.IBFTMessage](messages ...M) []Message {
	wrapped := make([]Message, 0, len(messages))

	for _, msg := range messages {
		//nolint:forcetypeassert // guarded by the types.IBFTMessage constraint
		wrapped = append(wrapped, any(msg).(Message))
	}

	return wrapped
}
