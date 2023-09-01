package types

import (
	"bytes"
	"google.golang.org/protobuf/proto"
)

func (x *MsgProposal) Payload() []byte {
	bz, _ := proto.Marshal(&MsgProposal{
		View:                   x.View,
		From:                   x.From,
		ProposedBlock:          x.ProposedBlock,
		ProposalHash:           x.ProposalHash,
		RoundChangeCertificate: x.RoundChangeCertificate,
	})

	return bz
}

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

func (x *MsgRoundChange) Payload() []byte {
	bz, _ := proto.Marshal(&MsgRoundChange{
		View:                        x.GetView(),
		From:                        x.GetFrom(),
		Signature:                   x.GetSignature(),
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

	proposalHash := pc.ProposalMessage.ProposalHash
	for _, msg := range pc.PrepareMessages {
		if !bytes.Equal(msg.ProposalHash, proposalHash) {
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
