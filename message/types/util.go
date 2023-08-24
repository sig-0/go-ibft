package types

import (
	"google.golang.org/protobuf/proto"
)

func (x *MsgPrepare) Payload() []byte {
	bz, _ := proto.Marshal(&MsgPrepare{
		View:         x.GetView(),
		From:         x.GetFrom(),
		ProposalHash: x.GetProposalHash(),
	})

	return bz
}

func (x *MsgCommit) Payload() []byte {
	bz, _ := proto.Marshal(&MsgCommit{
		View:         x.GetView(),
		From:         x.GetFrom(),
		ProposalHash: x.GetProposalHash(),
		CommitSeal:   x.GetCommitSeal(),
	})

	return bz
}

func (x *ProposedBlock) Bytes() []byte {
	bz, _ := proto.Marshal(x)

	return bz
}
