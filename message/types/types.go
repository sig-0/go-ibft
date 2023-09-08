package types

type FinalizedSeal struct {
	From, CommitSeal []byte
}

type FinalizedBlock struct {
	Block []byte
	Seals []FinalizedSeal
	Round uint64
}

// Msg is a convenience wrapper for consensus messages
type Msg interface {
	GetFrom() []byte
	GetSignature() []byte
	Payload() []byte
}

// MsgFeed provides an asynchronous way to receive consensus messages. In addition
// to listen for any type of message for any particular view, the higherRounds flag provides an option
// to receive messages from rounds higher than the round in provided view.
//
// CONTRACT: messages received by consuming the channel's callback are assumed to be valid:
//
// - any message has a valid view (matches the one provided)
//
// - any message has a valid signature (validator signed the message payload)
//
// - all messages are considered unique (there cannot be 2 or more messages with identical From fields)
type MsgFeed interface {
	Proposal(view *View, higherRounds bool) (<-chan func() []*MsgProposal, func())
	Prepare(view *View, higherRounds bool) (<-chan func() []*MsgPrepare, func())
	Commit(view *View, higherRounds bool) (<-chan func() []*MsgCommit, func())
	RoundChange(view *View, higherRounds bool) (<-chan func() []*MsgRoundChange, func())
}

type Signer interface {
	Sign([]byte) []byte
}

type SigRecover interface {
	From(data []byte, sig []byte) []byte
}
