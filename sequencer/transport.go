package sequencer

import (
	"github.com/sig-0/go-ibft/message"
)

// Transport is used by Validator to gossip consensus messages to other validators in the network
type Transport interface {
	// MulticastProposal gossips MsgProposal to other consensus peers
	MulticastProposal(msg *message.MsgProposal)

	// MulticastPrepare gossips MsgPrepare to other consensus peers
	MulticastPrepare(msg *message.MsgPrepare)

	// MulticastCommit gossips MsgCommit to other consensus peers
	MulticastCommit(msg *message.MsgCommit)

	// MulticastRoundChange gossips MsgRoundChange to other consensus peers
	MulticastRoundChange(msg *message.MsgRoundChange)
}
