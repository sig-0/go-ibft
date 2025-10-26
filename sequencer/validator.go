package sequencer

import (
	"context"

	"github.com/sig-0/go-ibft/message"
)

// Validator represents a unique consensus.go actor in the network whose role
// is to broadcast (signed) consensus.go messages and make proposals (if elected)
type Validator interface {
	message.Signer

	// BuildProposal returns this validator's proposal for given sequence
	BuildProposal(ctx context.Context, sequence uint64) ([]byte, error)
}

type ProposerAlgo interface {
	GetProposer(ctx context.Context, sequence, round uint64) ([]byte, error)
}
