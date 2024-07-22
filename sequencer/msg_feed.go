package sequencer

import (
	"github.com/sig-0/go-ibft/message/types"
)

// MsgFeed provides Sequencer an asynchronous way to receive consensus messages. In addition to
// listening for any type of message in any particular view, the higherRounds flag provides an option
// to include messages from higher rounds as well.
//
// CONTRACT:
//
// 1. any message is valid:
//   - no required fields missing (Sender, Signature, getView)
//   - signature is valid [ recover(ibftMsg.Payload, Signature) == Sender ]
//
// 2. all messages are considered unique (there cannot be 2 or more messages from the same sender)
type MsgFeed interface {
	// ProposalMessages returns the MsgProposal subscription for given view(s)
	ProposalMessages(view *types.View, higherRounds bool) (types.Subscription[*types.MsgProposal], func())

	// PrepareMessages returns the MsgPrepare subscription for given view(s)
	PrepareMessages(view *types.View, higherRounds bool) (types.Subscription[*types.MsgPrepare], func())

	// CommitMessages returns the MsgCommit subscription for given view(s)
	CommitMessages(view *types.View, higherRounds bool) (types.Subscription[*types.MsgCommit], func())

	// RoundChangeMessages returns the MsgRoundChange subscription for given view(s)
	RoundChangeMessages(view *types.View, higherRounds bool) (types.Subscription[*types.MsgRoundChange], func())
}
