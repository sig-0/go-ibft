package message

import (
	"golang.org/x/crypto/sha3"
	"google.golang.org/protobuf/proto"
)

type message interface {
	GetSender() []byte
	GetSequence() uint64
	GetSignature() []byte
	Payload() []byte
}

type Deriver interface {
	DeriveSender(digest, signature []byte) ([]byte, error)
}

type SignatureVerifier interface {
	Verify(sender, digest, signature []byte) error
}

type Signer interface {
	Sign([]byte) []byte
}

func Sign(msg message, signer Signer) []byte {
	return signer.Sign(keccak(msg.Payload()))
}

func VerifySignature(msg message, vrf SignatureVerifier) error {
	return vrf.Verify(msg.GetSender(), keccak(msg.Payload()), msg.GetSignature())
}

func keccak(input []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	defer hash.Reset()

	hash.Write(input)
	return hash.Sum(nil)
}

func GetProposalHash(pb *ProposedBlock) []byte {
	input := make([]byte, 0, len(pb.Block)+1)
	input = append(input, pb.Block...)
	input = append(input, byte(pb.Round))

	return keccak(input)
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

func (x *Proposal) IsMalformed() bool {
	return len(x.Sender) == 0 || len(x.Signature) == 0 || len(x.BlockHash) == 0 || x.ProposedBlock == nil
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

func (x *Prepare) IsMalformed() bool {
	return len(x.Sender) == 0 || len(x.Signature) == 0 || len(x.BlockHash) == 0
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

func (x *Commit) IsMalformed() bool {
	return len(x.Sender) == 0 || len(x.Signature) == 0 || len(x.BlockHash) == 0 || len(x.CommitSeal) == 0
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

func (x *RoundChange) IsMalformed() bool {
	return len(x.Sender) == 0 || len(x.Signature) == 0
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
