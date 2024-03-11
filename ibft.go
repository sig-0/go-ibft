package ibft

import "github.com/madz-lab/go-ibft/message/types"

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

	// Keccak is used to obtain the Keccak encoding of arbitrary data
	Keccak interface {
		// Hash returns the Keccak encoding of given input
		Hash(data []byte) []byte
	}

	// Quorum is used to check whether the protocol reached consensus during a particular step
	Quorum interface {
		// HasQuorum returns true if messages accumulate consensus for a particular sequence
		HasQuorum(messages []types.Message) bool
	}

	// Transport is used to gossip a consensus message to the network
	Transport[M types.IBFTMessage] interface {
		// Multicast gossips a consensus message to the network
		Multicast(M)
	}
)

type (
	SignerFn                         func([]byte) []byte
	KeccakFn                         func([]byte) []byte
	QuorumFn                         func([]types.Message) bool
	TransportFn[M types.IBFTMessage] func(M)
)

func (f SignerFn) Sign(digest []byte) []byte {
	return f(digest)
}

func (f QuorumFn) HasQuorum(messages []types.Message) bool {
	return f(messages)
}

func (f KeccakFn) Hash(data []byte) []byte {
	return f(data)
}

func (f TransportFn[M]) Multicast(msg M) {
	f(msg)
}
