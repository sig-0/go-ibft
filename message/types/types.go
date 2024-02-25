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

// FinalizedSeal is proof that a validator committed to a specific proposal
type FinalizedSeal struct {
	From, CommitSeal []byte
}

// FinalizedProposal is a consensus verified proposal of some sequence
type FinalizedProposal struct {
	// proposal that was finalized
	Proposal []byte

	// seals of validators who committed to this proposal
	Seals []FinalizedSeal

	// round in which the proposal was finalized
	Round uint64
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

func (rcc *RoundChangeCertificate) HighestRoundBlock() ([]byte, uint64) {
	roundsAndPreparedBlocks := make(map[uint64][]byte)

	for _, msg := range rcc.Messages {
		pb := msg.LatestPreparedProposedBlock
		pc := msg.LatestPreparedCertificate

		if pb == nil || pc == nil {
			continue
		}

		roundsAndPreparedBlocks[pc.ProposalMessage.View.Round] = pb.Block
	}

	if len(roundsAndPreparedBlocks) == 0 {
		return nil, 0
	}

	var (
		maxRound      uint64
		maxRoundBlock []byte
	)

	for round, block := range roundsAndPreparedBlocks {
		if round >= maxRound {
			maxRound = round
			maxRoundBlock = block
		}
	}

	return maxRoundBlock, maxRound
}

func (rcc *RoundChangeCertificate) HighestRoundBlockHash() ([]byte, uint64) {
	roundsAndPreparedBlockHashes := make(map[uint64][]byte)

	for _, msg := range rcc.Messages {
		pc := msg.LatestPreparedCertificate
		if pc == nil {
			continue
		}

		roundsAndPreparedBlockHashes[pc.ProposalMessage.View.Round] = pc.ProposalMessage.BlockHash
	}

	if len(roundsAndPreparedBlockHashes) == 0 {
		return nil, 0
	}

	var (
		maxRound             uint64
		maxRoundProposalHash []byte
	)

	for round, proposalHash := range roundsAndPreparedBlockHashes {
		if round >= maxRound {
			maxRound = round
			maxRoundProposalHash = proposalHash
		}
	}

	return maxRoundProposalHash, maxRound
}
