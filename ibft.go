package ibft

import "github.com/sig-0/go-ibft/message/types"

/*** External functionalities required by the IBFT 2.0 block finalization algorithm	***/

type (
	// Validator represents a unique consensus actor in the network.
	// Its role in the protocol is to broadcast (signed) consensus messages and make proposals (if elected)
	Validator interface {
		Signer

		// ID returns the network-unique ID of this validator
		ID() []byte

		// BuildProposal returns this validator's proposal for given sequence
		BuildProposal(sequence uint64) []byte

		// IsValidProposal checks if the provided proposal is valid for given sequence
		IsValidProposal(proposal []byte, sequence uint64) bool
	}

	ValidatorSet interface {
		// IsValidator checks if id is part of consensus for given sequence
		IsValidator(id []byte, sequence uint64) bool

		// IsProposer asserts if id is the elected proposer for given sequence and round
		IsProposer(id []byte, sequence, round uint64) bool
	}

	// Signer is used to generate a sender-unique signature of a keccak hash
	Signer interface {
		// Sign computes the signature of given digest
		Sign(digest []byte) []byte
	}

	// SigVerifier authenticates consensus messages gossiped in the network
	SigVerifier interface {
		// Verify checks if the signature of the message is valid
		Verify(signer, digest, signature []byte) error
	}

	// Keccak is used to obtain the Keccak encoding of arbitrary data
	Keccak interface {
		// Hash returns the Keccak encoding of given input
		Hash(data []byte) []byte
	}

	// Quorum is used to check if provided messages satisfy a consensus in the network
	Quorum interface {
		// HasQuorum returns true if messages accumulate consensus for a particular sequence
		HasQuorum(messages []types.Message) bool
	}

	// Transport is used to gossip consensus messages to the network
	Transport[M types.IBFTMessage] interface {
		// Multicast gossips a consensus message to the network
		Multicast(M)
	}
)

type (
	SignerFn                         func([]byte) []byte
	SigVerifierFn                    func([]byte, []byte, []byte) error
	KeccakFn                         func([]byte) []byte
	QuorumFn                         func([]types.Message) bool
	TransportFn[M types.IBFTMessage] func(M)
)

func (f SignerFn) Sign(digest []byte) []byte {
	return f(digest)
}

func (f SigVerifierFn) Verify(id, digest, sig []byte) error {
	return f(id, digest, sig)
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

type MsgTransport struct {
	Proposal    Transport[*types.MsgProposal]
	Prepare     Transport[*types.MsgPrepare]
	Commit      Transport[*types.MsgCommit]
	RoundChange Transport[*types.MsgRoundChange]
}

func NewMsgTransport(
	proposal func(*types.MsgProposal),
	prepare func(*types.MsgPrepare),
	commit func(*types.MsgCommit),
	roundChange func(change *types.MsgRoundChange),
) MsgTransport {
	return MsgTransport{
		Proposal:    TransportFn[*types.MsgProposal](proposal),
		Prepare:     TransportFn[*types.MsgPrepare](prepare),
		Commit:      TransportFn[*types.MsgCommit](commit),
		RoundChange: TransportFn[*types.MsgRoundChange](roundChange),
	}
}

func (t MsgTransport) MulticastProposal(msg *types.MsgProposal) {
	t.Proposal.Multicast(msg)
}

func (t MsgTransport) MulticastPrepare(msg *types.MsgPrepare) {
	t.Prepare.Multicast(msg)
}

func (t MsgTransport) MulticastCommit(msg *types.MsgCommit) {
	t.Commit.Multicast(msg)
}

func (t MsgTransport) MulticastRoundChange(msg *types.MsgRoundChange) {
	t.RoundChange.Multicast(msg)
}
