package types

import (
	"bytes"
	"google.golang.org/protobuf/proto"
)

type FinalizedSeal struct {
	From, CommitSeal []byte
}

type FinalizedBlock struct {
	Block []byte
	Seals []FinalizedSeal
	Round uint64
}

func (x *MsgProposal) Bytes() []byte {
	bz, _ := proto.Marshal(&MsgProposal{
		View:                   x.View,
		From:                   x.From,
		Signature:              x.Signature,
		ProposedBlock:          x.ProposedBlock,
		BlockHash:              x.BlockHash,
		RoundChangeCertificate: x.RoundChangeCertificate,
	})

	return bz
}

func (x *MsgProposal) Payload() []byte {
	bz, _ := proto.Marshal(&MsgProposal{
		View:                   x.View,
		From:                   x.From,
		ProposedBlock:          x.ProposedBlock,
		BlockHash:              x.BlockHash,
		RoundChangeCertificate: x.RoundChangeCertificate,
	})

	return bz
}

func (x *MsgPrepare) Bytes() []byte {
	bz, _ := proto.Marshal(&MsgPrepare{
		View:      x.View,
		From:      x.From,
		Signature: x.Signature,
		BlockHash: x.BlockHash,
	})

	return bz
}

func (x *MsgPrepare) Payload() []byte {
	bz, _ := proto.Marshal(&MsgPrepare{
		View:      x.View,
		From:      x.From,
		BlockHash: x.BlockHash,
	})

	return bz
}

func (x *MsgCommit) Bytes() []byte {
	bz, _ := proto.Marshal(&MsgCommit{
		View:       x.View,
		From:       x.From,
		Signature:  x.Signature,
		BlockHash:  x.BlockHash,
		CommitSeal: x.CommitSeal,
	})

	return bz
}

func (x *MsgCommit) Payload() []byte {
	bz, _ := proto.Marshal(&MsgCommit{
		View:       x.View,
		From:       x.From,
		BlockHash:  x.BlockHash,
		CommitSeal: x.GetCommitSeal(),
	})

	return bz
}

func (x *MsgRoundChange) Bytes() []byte {
	bz, _ := proto.Marshal(&MsgRoundChange{
		View:                        x.View,
		From:                        x.From,
		Signature:                   x.Signature,
		LatestPreparedProposedBlock: x.LatestPreparedProposedBlock,
		LatestPreparedCertificate:   x.LatestPreparedCertificate,
	})

	return bz
}

func (x *MsgRoundChange) Payload() []byte {
	bz, _ := proto.Marshal(&MsgRoundChange{
		View:                        x.GetView(),
		From:                        x.GetFrom(),
		LatestPreparedProposedBlock: x.GetLatestPreparedProposedBlock(),
		LatestPreparedCertificate:   x.GetLatestPreparedCertificate(),
	})

	return bz
}

func (x *ProposedBlock) Bytes() []byte {
	bz, _ := proto.Marshal(x)

	return bz
}

func (pc *PreparedCertificate) IsValid(view *View) bool {
	if pc.ProposalMessage == nil || pc.PrepareMessages == nil {
		return false
	}

	if pc.ProposalMessage.View.Sequence != view.Sequence {
		return false
	}

	if pc.ProposalMessage.View.Round >= view.Round {
		return false
	}

	sequence, round := pc.ProposalMessage.View.Sequence, pc.ProposalMessage.View.Round
	for _, msg := range pc.PrepareMessages {
		if msg.View.Sequence != sequence {
			return false
		}

		if msg.View.Round != round {
			return false
		}
	}

	for _, msg := range pc.PrepareMessages {
		if !bytes.Equal(msg.BlockHash, pc.ProposalMessage.BlockHash) {
			return false
		}
	}

	senders := map[string]struct{}{string(pc.ProposalMessage.From): {}}
	for _, msg := range pc.PrepareMessages {
		senders[string(msg.From)] = struct{}{}
	}

	if len(senders) != 1+len(pc.PrepareMessages) {
		return false
	}

	return true
}

func (rcc *RoundChangeCertificate) IsValid(view *View) bool {
	if len(rcc.Messages) == 0 {
		return false
	}

	for _, msg := range rcc.Messages {
		if msg.View.Sequence != view.Sequence {
			return false
		}

		if msg.View.Round != view.Round {
			return false
		}
	}

	senders := make(map[string]struct{})
	for _, msg := range rcc.Messages {
		senders[string(msg.From)] = struct{}{}
	}

	if len(senders) != len(rcc.Messages) {
		return false
	}

	return true
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
