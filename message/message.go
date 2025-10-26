package message

import (
	"bytes"
	"errors"
	"fmt"

	"golang.org/x/crypto/sha3"
	"google.golang.org/protobuf/proto"
)

var (
	ErrMissingField = errors.New("missing field in message")
)

type Message interface {
	GetSender() []byte
	GetSequence() uint64
	GetSignature() []byte
	Payload() []byte
}

type Signer interface {
	Address() []byte
	Sign([]byte) []byte
}

type Deriver interface {
	DeriveSender(digest, signature []byte) ([]byte, error)
}

func SignMessage(msg Message, signer Signer) []byte {
	return signer.Sign(keccak(msg.Payload()))
}

func VerifyMessage(msg Message, d Deriver) error {
	if err := verifyFields(msg); err != nil {
		return err
	}

	digest := keccak(msg.Payload())
	sender, err := d.DeriveSender(digest, msg.GetSignature())
	if err != nil {
		return err
	}

	if !bytes.Equal(sender, msg.GetSender()) {
		return fmt.Errorf("sender mismatch: %s != %s", string(sender), string(msg.GetSender()))
	}

	return nil
}

func verifyFields(msg Message) error {
	switch msg := msg.(type) {
	case *Proposal:
		if len(msg.Sender) == 0 {
			return fmt.Errorf("%w: no sender", ErrMissingField)
		}

		if len(msg.Signature) == 0 {
			return fmt.Errorf("%w: no signature", ErrMissingField)
		}

		if len(msg.BlockHash) == 0 {
			return fmt.Errorf("%w: no block hash", ErrMissingField)
		}

		if msg.ProposedBlock == nil {
			return fmt.Errorf("%w: no proposed block", ErrMissingField)
		}

		return nil
	case *Prepare:
		if len(msg.Sender) == 0 {
			return fmt.Errorf("%w: no sender", ErrMissingField)
		}

		if len(msg.Signature) == 0 {
			return fmt.Errorf("%w: no signature", ErrMissingField)
		}

		if len(msg.BlockHash) == 0 {
			return fmt.Errorf("%w: no block hash", ErrMissingField)
		}

		return nil
	case *RoundChange:
		if len(msg.Sender) == 0 {
			return fmt.Errorf("%w: no sender", ErrMissingField)
		}

		if len(msg.Signature) == 0 {
			return fmt.Errorf("%w: no signature", ErrMissingField)
		}

		return nil
	case *Commit:
		if len(msg.Sender) == 0 {
			return fmt.Errorf("%w: no sender", ErrMissingField)
		}

		if len(msg.Signature) == 0 {
			return fmt.Errorf("%w: no signature", ErrMissingField)
		}

		if len(msg.BlockHash) == 0 {
			return fmt.Errorf("%w: no block hash", ErrMissingField)
		}

		if len(msg.CommitSeal) == 0 {
			return fmt.Errorf("%w: no commit seal", ErrMissingField)
		}

		return nil
	default:
		return errors.New("unknown message type")
	}
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
