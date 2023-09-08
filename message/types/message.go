package types

import "google.golang.org/protobuf/proto"

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
