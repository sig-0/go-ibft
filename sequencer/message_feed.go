package sequencer

import "github.com/madz-lab/go-ibft/message/types"

// MessageFeed provides an asynchronous way to receive consensus messages. In addition
// to listen for any type of message for any particular view, the higherRounds flag provides an option
// to include messages from rounds higher than provided. Given some view and higherRounds set to true,
// messages received can originate from future rounds.
//
// CONTRACT: messages received by consuming the channel's callback are assumed to be valid:
//
// - any message has a valid view (matches the one previously provided)
//
// - any message has a valid signature (validator signed the message payload)
//
// - all messages are considered unique (there cannot be 2 or more messages with identical From fields)
type MessageFeed interface {
	SubscribeToProposalMessages(view *types.View, futureRounds bool) (<-chan func() []*types.MsgProposal, func())
	SubscribeToPrepareMessages(view *types.View, futureRounds bool) (<-chan func() []*types.MsgPrepare, func())
	SubscribeToCommitMessages(view *types.View, futureRounds bool) (<-chan func() []*types.MsgCommit, func())
	SubscribeToRoundChangeMessages(view *types.View, futureRounds bool) (<-chan func() []*types.MsgRoundChange, func())
}
