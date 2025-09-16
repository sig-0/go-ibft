package message

import (
	"google.golang.org/protobuf/proto"
)

type message interface {
	GetSender() []byte
	GetSequence() uint64
	GetRound() uint64
}

func (x *Proposal) Payload() []byte {
	xx := &Proposal{
		Sequence:               x.Sequence,
		Round:                  x.Round,
		Sender:                 x.Sender,
		ProposedBlock:          x.ProposedBlock,
		BlockHash:              x.BlockHash,
		RoundChangeCertificate: x.RoundChangeCertificate,
	}

	payload, _ := proto.Marshal(xx) //nolint:errcheck //proto
	return payload
}

func (x *Prepare) Payload() []byte {
	xx := &Prepare{
		Sequence:  x.Sequence,
		Round:     x.Round,
		Sender:    x.Sender,
		BlockHash: x.BlockHash,
	}

	payload, _ := proto.Marshal(xx) //nolint:errcheck //proto
	return payload
}

func (x *Commit) Payload() []byte {
	xx := &Commit{
		Sequence:   x.Sequence,
		Round:      x.Round,
		Sender:     x.Sender,
		BlockHash:  x.BlockHash,
		CommitSeal: x.CommitSeal,
	}

	payload, _ := proto.Marshal(xx) //nolint:errcheck //proto
	return payload
}

func (x *RoundChange) Payload() []byte {
	xx := &RoundChange{
		Sequence:                    x.Sequence,
		Round:                       x.Round,
		Sender:                      x.Sender,
		LatestPreparedProposedBlock: x.LatestPreparedProposedBlock,
		LatestPreparedCertificate:   x.LatestPreparedCertificate,
	}

	payload, _ := proto.Marshal(xx) //nolint:errcheck //proto
	return payload
}

func (x *ProposedBlock) Bytes() []byte {
	bz, _ := proto.Marshal(x) //nolint:errcheck //proto
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

		roundsAndPreparedBlocks[pc.ProposalMessage.Round] = pb.Block
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

		roundsAndPreparedBlockHashes[pc.ProposalMessage.Round] = pc.ProposalMessage.BlockHash
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
