package types

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"
)

var (
	ErrInvalidMessage = errors.New("invalid consensus message")
)

// IBFTMessage defines the 4 message types used by the IBFT 2.0 protocol
// to reach network-wide consensus on some proposal for a particular sequence
type IBFTMessage interface {
	*MsgProposal | *MsgPrepare | *MsgCommit | *MsgRoundChange
}

// Message is an opaque wrapper for the IBFT consensus messages. See IBFTMessage for concrete type definitions
type Message interface {
	// Validate returns an errors if the Message is malformed
	Validate() error

	// Sequence in which this Message was sent
	Sequence() uint64

	// Round in which this Message was sent
	Round() uint64

	// Sender returns the sender ID associated with this Message
	Sender() []byte

	// Signature returns the signature of this Message
	Signature() []byte

	// Payload returns the byte content of this Message (signature excluded)
	Payload() []byte

	// Bytes returns the byte content of this Message (signature included)
	Bytes() []byte
}

// WrapMessages wraps concrete message types into Message type
func WrapMessages[M IBFTMessage](messages ...M) []Message {
	wrapped := make([]Message, 0, len(messages))

	for _, msg := range messages {
		//nolint:forcetypeassert // guarded by the types.IBFTMessage constraint
		wrapped = append(wrapped, any(msg).(Message))
	}

	return wrapped
}

func (x *MsgMetadata) Validate() error {
	if x == nil {
		return fmt.Errorf("%w: missing metadata", ErrInvalidMessage)
	}

	if x.View == nil {
		return fmt.Errorf("%w: missing view", ErrInvalidMessage)
	}

	if x.Sender == nil {
		return fmt.Errorf("%w: missing sender", ErrInvalidMessage)
	}

	if x.Signature == nil {
		return fmt.Errorf("%w: missing signature", ErrInvalidMessage)
	}

	return nil
}

// MsgProposal

func (x *MsgProposal) Sequence() uint64 {
	return x.Metadata.View.Sequence
}

func (x *MsgProposal) Round() uint64 {
	return x.Metadata.View.Round
}

func (x *MsgProposal) Sender() []byte {
	return x.Metadata.Sender
}

func (x *MsgProposal) Signature() []byte {
	return x.Metadata.Signature
}

func (x *MsgProposal) Validate() error {
	if err := x.Metadata.Validate(); err != nil {
		return err
	}

	if x.BlockHash == nil {
		return fmt.Errorf("%w: missing block hash", ErrInvalidMessage)
	}

	if x.ProposedBlock == nil {
		return fmt.Errorf("%w: missing proposed block", ErrInvalidMessage)
	}

	return nil
}

func (x *MsgProposal) Payload() []byte {
	xx := &MsgProposal{
		Metadata: &MsgMetadata{
			View:   x.Metadata.View,
			Sender: x.Metadata.Sender,
		},
		ProposedBlock:          x.ProposedBlock,
		BlockHash:              x.BlockHash,
		RoundChangeCertificate: x.RoundChangeCertificate,
	}

	bz, err := proto.Marshal(xx)
	if err != nil {
		panic(fmt.Errorf("failed to generate payload of MsgProposal: %w", err).Error())
	}

	return bz
}

func (x *MsgProposal) Bytes() []byte {
	bz, err := proto.Marshal(x)
	if err != nil {
		panic(fmt.Errorf("failed to marshal MsgProposal: %w", err).Error())
	}

	return bz
}

// MsgPrepare

func (x *MsgPrepare) Sequence() uint64 {
	return x.Metadata.View.Sequence
}

func (x *MsgPrepare) Round() uint64 {
	return x.Metadata.View.Round
}

func (x *MsgPrepare) Sender() []byte {
	return x.Metadata.Sender
}

func (x *MsgPrepare) Signature() []byte {
	return x.Metadata.Signature
}

func (x *MsgPrepare) Validate() error {
	if err := x.Metadata.Validate(); err != nil {
		return err
	}

	if x.BlockHash == nil {
		return fmt.Errorf("%w: missing block hash", ErrInvalidMessage)
	}

	return nil
}

func (x *MsgPrepare) Bytes() []byte {
	bz, err := proto.Marshal(x)
	if err != nil {
		panic(fmt.Errorf("failed to marshal MsgPrepare: %w", err).Error())
	}

	return bz
}

func (x *MsgPrepare) Payload() []byte {
	xx := &MsgPrepare{
		Metadata: &MsgMetadata{
			View:   x.Metadata.View,
			Sender: x.Metadata.Sender,
		},
		BlockHash: x.BlockHash,
	}

	bz, err := proto.Marshal(xx)
	if err != nil {
		panic(fmt.Errorf("failed to generate payload of MsgPrepare: %w", err).Error())
	}

	return bz
}

// MsgCommit

func (x *MsgCommit) Sequence() uint64 {
	return x.Metadata.View.Sequence
}

func (x *MsgCommit) Round() uint64 {
	return x.Metadata.View.Round
}

func (x *MsgCommit) Sender() []byte {
	return x.Metadata.Sender
}

func (x *MsgCommit) Signature() []byte {
	return x.Metadata.Signature
}

func (x *MsgCommit) Validate() error {
	if err := x.Metadata.Validate(); err != nil {
		return err
	}

	if x.BlockHash == nil {
		return fmt.Errorf("%w: missing block hash", ErrInvalidMessage)
	}

	if x.CommitSeal == nil {
		return fmt.Errorf("%w: missing commit seal", ErrInvalidMessage)
	}

	return nil
}

func (x *MsgCommit) Bytes() []byte {
	bz, err := proto.Marshal(x)
	if err != nil {
		panic(fmt.Errorf("failed to marshal MsgCommit: %w", err).Error())
	}

	return bz
}

func (x *MsgCommit) Payload() []byte {
	xx := &MsgCommit{
		Metadata: &MsgMetadata{
			View:   x.Metadata.View,
			Sender: x.Metadata.Sender,
		},
		BlockHash:  x.BlockHash,
		CommitSeal: x.CommitSeal,
	}

	bz, err := proto.Marshal(xx)
	if err != nil {
		panic(fmt.Errorf("failed to generate payload of MsgCommit: %w", err).Error())
	}

	return bz
}

// MsgRoundChange

func (x *MsgRoundChange) Sequence() uint64 {
	return x.Metadata.View.Sequence
}

func (x *MsgRoundChange) Round() uint64 {
	return x.Metadata.View.Round
}

func (x *MsgRoundChange) Sender() []byte {
	return x.Metadata.Sender
}

func (x *MsgRoundChange) Signature() []byte {
	return x.Metadata.Signature
}

func (x *MsgRoundChange) Validate() error {
	return x.Metadata.Validate()
}

func (x *MsgRoundChange) Bytes() []byte {
	bz, err := proto.Marshal(x)
	if err != nil {
		panic(fmt.Errorf("failed to marshal MsgRoundChange: %w", err).Error())
	}

	return bz
}

func (x *MsgRoundChange) Payload() []byte {
	xx := &MsgRoundChange{
		Metadata: &MsgMetadata{
			View:   x.Metadata.View,
			Sender: x.Metadata.Sender,
		},
		LatestPreparedProposedBlock: x.LatestPreparedProposedBlock,
		LatestPreparedCertificate:   x.LatestPreparedCertificate,
	}

	bz, err := proto.Marshal(xx)
	if err != nil {
		panic(fmt.Errorf("failed to genereate payload of MsgRoundChange: %w", err).Error())
	}

	return bz
}

func (x *ProposedBlock) Bytes() []byte {
	bz, err := proto.Marshal(x)
	if err != nil {
		panic(fmt.Errorf("failed to marshal ProposedBlock: %w", err).Error())
	}

	return bz
}
