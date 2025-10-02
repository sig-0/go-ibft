package sequencer

import (
	"github.com/sig-0/go-ibft/message"
)

// Transport is used by Validator to gossip consensus.go messages to other validators in the network
type Transport interface {
	// MulticastProposal gossips MsgProposal to other consensus.go peers
	MulticastProposal(*message.Proposal)

	// MulticastPrepare gossips MsgPrepare to other consensus.go peers
	MulticastPrepare(*message.Prepare)

	// MulticastCommit gossips MsgCommit to other consensus.go peers
	MulticastCommit(*message.Commit)

	// MulticastRoundChange gossips MsgRoundChange to other consensus.go peers
	MulticastRoundChange(*message.RoundChange)
}
