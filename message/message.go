package message

import (
	"google.golang.org/protobuf/proto"
)

// IBFTMessage defines the 4 message types used in the IBFT 2.0 protocol
// to reach network-wide consensus on some proposal for a particular sequence (height)
type IBFTMessage interface {
	*MsgProposal | *MsgPrepare | *MsgCommit | *MsgRoundChange
}

// Message is an opaque wrapper for the IBFT consensus messages. See IBFTMessage for concrete type definitions
type Message interface {
	GetInfo() *MsgInfo
}

type Signer interface {
	Address() []byte

	// Sign computes the signature of given digest
	Sign(digest []byte) []byte
}

type SignatureVerifier interface {
	// Verify checks if the signature of the message is valid
	Verify(signer, digest, signature []byte) error
}

type Keccak interface {
	// Hash returns the Keccak encoding of given input
	Hash(data []byte) []byte
}

// Transport is used to gossip consensus messages to the network
type Transport interface {
	MulticastProposal(msg *MsgProposal)
	MulticastPrepare(msg *MsgPrepare)
	MulticastCommit(msg *MsgCommit)
	MulticastRoundChange(msg *MsgRoundChange)
}

// Feed provides Sequencer an asynchronous way to receive consensus messages. In addition to
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
type Feed interface {
	// SubscribeProposal returns the MsgProposal subscription for given view(s)
	SubscribeProposal(view *View, higherRounds bool) (Subscription[*MsgProposal], func())

	// SubscribePrepare returns the MsgPrepare subscription for given view(s)
	SubscribePrepare(view *View, higherRounds bool) (Subscription[*MsgPrepare], func())

	// SubscribeCommit returns the MsgCommit subscription for given view(s)
	SubscribeCommit(view *View, higherRounds bool) (Subscription[*MsgCommit], func())

	// SubscribeRoundChange returns the MsgRoundChange subscription for given view(s)
	SubscribeRoundChange(view *View, higherRounds bool) (Subscription[*MsgRoundChange], func())
}

type (
	// MsgNotification is received from the subscription to indicate a new message
	MsgNotification[M IBFTMessage] interface {
		// Unwrap returns all messages that fit the subscription
		Unwrap() []M
	}

	MsgNotificationFn[M IBFTMessage] func() []M

	Subscription[M IBFTMessage] chan MsgNotification[M]
)

func (r MsgNotificationFn[M]) Unwrap() []M {
	return r()
}

// WrapMessages wraps concrete message types into Message type
func WrapMessages[M IBFTMessage](messages ...M) []Message {
	wrapped := make([]Message, 0, len(messages))
	for _, msg := range messages {
		wrapped = append(wrapped, Message(msg))
	}

	return wrapped
}

func SignMsg[M IBFTMessage](msg M, signer Signer) M {
	switch m := Message(msg).(type) {
	case *MsgProposal:
		m.Info.Signature = nil
		payload, _ := proto.Marshal(m)
		m.Info.Signature = signer.Sign(payload)
	case *MsgPrepare:
		m.Info.Signature = nil
		payload, _ := proto.Marshal(m)
		m.Info.Signature = signer.Sign(payload)
	case *MsgCommit:
		m.Info.Signature = nil
		payload, _ := proto.Marshal(m)
		m.Info.Signature = signer.Sign(payload)
	case *MsgRoundChange:
		m.Info.Signature = nil
		payload, _ := proto.Marshal(m)
		m.Info.Signature = signer.Sign(payload)
	}

	return msg
}

func (x *ProposedBlock) Bytes() []byte {
	bz, _ := proto.Marshal(x)
	return bz
}

func (rcc *RoundChangeCertificate) HighestRoundBlock() ([]byte, uint64) {
	roundsAndPreparedBlocks := make(map[uint64][]byte)
	for _, msg := range rcc.Messages {
		pb := msg.LatestPreparedProposedBlock
		pc := msg.LatestPreparedCertificate

		if pb == nil || pc == nil {
			continue
		}

		roundsAndPreparedBlocks[pc.ProposalMessage.Info.View.Round] = pb.Block
	}

	if len(roundsAndPreparedBlocks) == 0 {
		return nil, 0
	}

	var (
		highestRound      uint64
		highestRoundBlock []byte
	)

	for round, block := range roundsAndPreparedBlocks {
		if round >= highestRound {
			highestRound = round
			highestRoundBlock = block
		}
	}

	return highestRoundBlock, highestRound
}

func (rcc *RoundChangeCertificate) HighestRoundBlockHash() ([]byte, uint64) {
	roundsAndPreparedBlockHashes := make(map[uint64][]byte)
	for _, msg := range rcc.Messages {
		pc := msg.LatestPreparedCertificate
		if pc == nil {
			continue
		}

		roundsAndPreparedBlockHashes[pc.ProposalMessage.Info.View.Round] = pc.ProposalMessage.BlockHash
	}

	if len(roundsAndPreparedBlockHashes) == 0 {
		return nil, 0
	}

	var (
		highestRound          uint64
		highestRoundBlockHash []byte
	)

	for round, proposalHash := range roundsAndPreparedBlockHashes {
		if round >= highestRound {
			highestRound = round
			highestRoundBlockHash = proposalHash
		}
	}

	return highestRoundBlockHash, highestRound
}
