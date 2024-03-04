package types

import (
	"fmt"
	"google.golang.org/protobuf/proto"
)

// IBFTMessage defines the 4 message types used by the IBFT 2.0 protocol
// to reach network-wide consensus on some proposal for a particular sequence
type IBFTMessage interface {
	*MsgProposal | *MsgPrepare | *MsgCommit | *MsgRoundChange
}

func (x *MsgProposal) Bytes() []byte {
	bz, err := proto.Marshal(x)
	if err != nil {
		panic(fmt.Errorf("failed to marshal MsgProposal: %w", err).Error())
	}

	return bz
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

func (x *MsgProposal) GetSequence() uint64 {
	return x.View.Sequence
}

func (x *MsgProposal) GetRound() uint64 {
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

func (x *MsgPrepare) GetSequence() uint64 {
	return x.View.Sequence
}

func (x *MsgPrepare) GetRound() uint64 {
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

func (x *MsgCommit) GetSequence() uint64 {
	return x.View.Sequence
}

func (x *MsgCommit) GetRound() uint64 {
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

func (x *MsgRoundChange) GetSequence() uint64 {
	return x.View.Sequence
}

func (x *MsgRoundChange) GetRound() uint64 {
	return x.View.Round
}

func (x *ProposedBlock) Bytes() []byte {
	bz, err := proto.Marshal(x)
	if err != nil {
		panic(fmt.Errorf("failed to marshal ProposedBlock: %w", err).Error())
	}

	return bz
}
