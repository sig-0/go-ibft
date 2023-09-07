package types

import "google.golang.org/protobuf/proto"

type FinalizedSeal struct {
	From, CommitSeal []byte
}

type FinalizedBlock struct {
	Block []byte
	Seals []FinalizedSeal
	Round uint64
}

type Msg interface {
	GetFrom() []byte
	GetSignature() []byte
	Payload() []byte
}

type SigRecover interface {
	From(data []byte, sig []byte) []byte
}

type Signer interface {
	Sign([]byte) []byte
}

type ibft interface {
	MsgProposal | MsgPrepare | MsgCommit | MsgRoundChange
}

func Filter[M ibft](messages []*M, filter func(*M) bool) []*M {
	filtered := make([]*M, 0, len(messages))
	for _, msg := range messages {
		if filter(msg) {
			filtered = append(filtered, msg)
		}
	}

	return filtered
}

func ToMsg[M ibft](messages []*M) []Msg {
	wrapped := make([]Msg, 0, len(messages))
	for _, msg := range messages {
		wrapped = append(wrapped, any(msg).(Msg))
	}

	return wrapped
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
		CommitSeal: x.CommitSeal,
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
		View:                        x.View,
		From:                        x.From,
		LatestPreparedProposedBlock: x.LatestPreparedProposedBlock,
		LatestPreparedCertificate:   x.LatestPreparedCertificate,
	})

	return bz
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
