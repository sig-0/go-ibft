package sequencer

import (
	"context"

	"github.com/sig-0/go-ibft/message"
)

// Transport is used by Validator to gossip consensus.go messages to other validators in the network
type Transport interface {
	// MulticastProposal gossips MsgProposal to other consensus.go peers
	MulticastProposal(ctx context.Context, msg *message.Proposal)

	// MulticastPrepare gossips MsgPrepare to other consensus.go peers
	MulticastPrepare(ctx context.Context, msg *message.Prepare)

	// MulticastCommit gossips MsgCommit to other consensus.go peers
	MulticastCommit(ctx context.Context, msg *message.Commit)

	// MulticastRoundChange gossips MsgRoundChange to other consensus.go peers
	MulticastRoundChange(ctx context.Context, msg *message.RoundChange)
}
