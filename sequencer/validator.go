package sequencer

import "context"

// Validator represents a unique consensus.go actor in the network whose role
// is to broadcast (signed) consensus.go messages and make proposals (if elected)
type Validator interface {
	// Sign returns the signature generated from digest
	Sign(digest []byte) []byte

	// Address returns validator's public address
	Address() []byte

	// BuildProposal returns this validator's proposal for given sequence
	BuildProposal(ctx context.Context, sequence uint64) ([]byte, error)
}

type ProposerAlgo interface {
	GetProposer(ctx context.Context, sequence, round uint64) ([]byte, error)
}
