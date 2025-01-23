package sequencer

import "github.com/sig-0/go-ibft/message"

// Validator represents a unique consensus actor in the network whose role
// is to broadcast (signed) consensus messages and make proposals (if elected)
type Validator interface {
	message.Signer

	// Address returns validator's public address
	Address() []byte

	// BuildProposal returns this validator's proposal for given sequence
	BuildProposal(sequence uint64) []byte

	// IsValidProposal checks if the provided proposal is valid for given sequence
	IsValidProposal(proposal []byte, sequence uint64) bool
}

type ValidatorSet interface {
	// IsValidator checks if id is part of consensus for given sequence
	IsValidator(addr []byte, sequence uint64) bool

	// IsProposer asserts if id is the elected proposer for given sequence and round
	IsProposer(addr []byte, sequence, round uint64) bool

	// HasQuorum returns true if messages accumulate consensus for a particular sequence
	HasQuorum(messages []message.Message) bool
}
