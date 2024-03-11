package types

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"
)

var (
	ErrMissingView          = errors.New("missing view field")
	ErrMissingSender        = errors.New("missing sender field")
	ErrMissingSignature     = errors.New("missing signature field")
	ErrMissingProposedBlock = errors.New("missing proposed block field")
	ErrMissingBlockHash     = errors.New("missing block hash field")
	ErrMissingCommitSeal    = errors.New("missing commit seal field")
)

// Message is an opaque wrapper for the IBFT consensus messages. See IBFTMessage for concrete type definitions
type Message interface {
	// Validate returns an errors if the Message is malformed
	Validate() error

	// GetSender returns the sender ID associated with this Message
	GetSender() []byte

	// GetSignature returns the signature of this Message
	GetSignature() []byte

	// Payload returns the byte content of this Message (signature excluded)
	Payload() []byte

	// Bytes returns the byte content of this Message (signature included)
	Bytes() []byte
}

// IBFTMessage defines the 4 message types used by the IBFT 2.0 protocol
// to reach network-wide consensus on some proposal for a particular sequence
type IBFTMessage interface {
	*MsgProposal | *MsgPrepare | *MsgCommit | *MsgRoundChange
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
	if x.View == nil {
		return ErrMissingView
	}

	if x.From == nil {
		return ErrMissingSender
	}

	if x.Signature == nil {
		return ErrMissingSignature
	}

	if x.BlockHash == nil {
		return ErrMissingBlockHash
	}

	if x.ProposedBlock == nil {
		return ErrMissingProposedBlock
	}

	return nil
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
		return ErrMissingView
	}

	if x.From == nil {
		return ErrMissingSender
	}

	if x.Signature == nil {
		return ErrMissingSignature
	}

	if x.BlockHash == nil {
		return ErrMissingBlockHash
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
		return ErrMissingView
	}

	if x.From == nil {
		return ErrMissingSender
	}

	if x.Signature == nil {
		return ErrMissingSignature
	}

	if x.BlockHash == nil {
		return ErrMissingBlockHash
	}

	if x.CommitSeal == nil {
		return ErrMissingCommitSeal
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
		return ErrMissingView
	}

	if x.From == nil {
		return ErrMissingSender
	}

	if x.Signature == nil {
		return ErrMissingSignature
	}

	return nil
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
