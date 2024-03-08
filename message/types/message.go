package types

import (
	"errors"
	"fmt"
	"google.golang.org/protobuf/proto"
)

// IBFTMessage defines the 4 message types used by the IBFT 2.0 protocol
// to reach network-wide consensus on some proposal for a particular sequence
type IBFTMessage interface {
	*MsgProposal | *MsgPrepare | *MsgCommit | *MsgRoundChange
}

// Message is an opaque wrapper for the IBFT consensus messages. See types.IBFTMessage for concrete type definitions
type Message interface {
	// Validate returns an errors if the Message is malformed
	Validate() error

	// GetSequence returns the sequence for which this Message was created
	GetSequence() uint64

	// GetRound returns the round in which this Message was created
	GetRound() uint64

	// GetSender returns the sender ID associated with this Message
	GetSender() []byte

	// GetSignature returns the signature of this Message
	GetSignature() []byte

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

func (x *MsgProposal) Validate() error {

	if x.BlockHash == nil {
		return errors.New("missing block hash")
	}

	return nil
}

func (x *MsgProposal) GetSequence() uint64 {
	return x.View.Sequence
}

func (x *MsgProposal) GetRound() uint64 {
	return x.View.Round
}

func (x *MsgProposal) Payload() []byte {
	bz, err := proto.Marshal(&MsgProposal{
		View:                   x.View,
		From:                   x.From,
		ProposedBlock:          x.ProposedBlock,
		BlockHash:              x.BlockHash,
		RoundChangeCertificate: x.RoundChangeCertificate,
	})
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

func (x *MsgPrepare) Validate() error {
	if x.View == nil {
		return errors.New("missing view")
	}

	if x.From == nil {
		return errors.New("missing sender")
	}

	if x.Signature == nil {
		return errors.New("missing signature")
	}

	if x.BlockHash == nil {
		return errors.New("missing block hash")
	}

	return nil
}

func (x *MsgPrepare) GetSequence() uint64 {
	return x.View.Sequence
}

func (x *MsgPrepare) GetRound() uint64 {
	return x.View.Round
}

func (x *MsgPrepare) Bytes() []byte {
	bz, err := proto.Marshal(x)
	if err != nil {
		panic(fmt.Errorf("failed to marshal MsgPrepare: %w", err).Error())
	}

	return bz
}

func (x *MsgPrepare) Payload() []byte {
	bz, err := proto.Marshal(&MsgPrepare{
		View:      x.View,
		From:      x.From,
		BlockHash: x.BlockHash,
	})
	if err != nil {
		panic(fmt.Errorf("failed to generate payload of MsgPrepare: %w", err).Error())
	}

	return bz
}

func (x *MsgCommit) Validate() error {
	if x.View == nil {
		return errors.New("missing view")
	}

	if x.From == nil {
		return errors.New("missing sender")
	}

	if x.Signature == nil {
		return errors.New("missing signature")
	}

	if x.BlockHash == nil {
		return errors.New("missing block hash")
	}

	if x.CommitSeal == nil {
		return errors.New("missing seal")
	}

	return nil
}

func (x *MsgCommit) GetSequence() uint64 {
	return x.View.Sequence
}

func (x *MsgCommit) GetRound() uint64 {
	return x.View.Round
}

func (x *MsgCommit) Bytes() []byte {
	bz, err := proto.Marshal(x)
	if err != nil {
		panic(fmt.Errorf("failed to marshal MsgCommit: %w", err).Error())
	}

	return bz
}

func (x *MsgCommit) Payload() []byte {
	bz, err := proto.Marshal(&MsgCommit{
		View:       x.View,
		From:       x.From,
		BlockHash:  x.BlockHash,
		CommitSeal: x.CommitSeal,
	})
	if err != nil {
		panic(fmt.Errorf("failed to generate payload of MsgCommit: %w", err).Error())
	}

	return bz
}

func (x *MsgRoundChange) Validate() error {
	if x.View == nil {
		return errors.New("missing view")
	}

	if x.From == nil {
		return errors.New("missing sender")
	}

	if x.Signature == nil {
		return errors.New("missing signature")
	}

	return nil
}

func (x *MsgRoundChange) GetSequence() uint64 {
	return x.View.Sequence
}

func (x *MsgRoundChange) GetRound() uint64 {
	return x.View.Round
}

func (x *MsgRoundChange) Bytes() []byte {
	bz, err := proto.Marshal(x)
	if err != nil {
		panic(fmt.Errorf("failed to marshal MsgRoundChange: %w", err).Error())
	}

	return bz
}

func (x *MsgRoundChange) Payload() []byte {
	bz, err := proto.Marshal(&MsgRoundChange{
		View:                        x.View,
		From:                        x.From,
		LatestPreparedProposedBlock: x.LatestPreparedProposedBlock,
		LatestPreparedCertificate:   x.LatestPreparedCertificate,
	})
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
