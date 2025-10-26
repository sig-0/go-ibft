package consensus

import (
	"context"

	"github.com/sig-0/go-ibft/message"
	"github.com/sig-0/go-ibft/sequencer"
)

type ValidatorSet interface {
	sequencer.ProposerAlgo

	GetValidators(ctx context.Context, sequence uint64) ([][]byte, error)
	CheckQuorum(ctx context.Context, sequence uint64, validators [][]byte) (bool, error)
}

type Verifier interface {
	VerifyProposal(ctx context.Context, sequence uint64, proposal []byte) error
}

type Consensus struct {
	vs                ValidatorSet
	vrf               Verifier
	deriver           message.Deriver
	currentValidators map[string]struct{}
}

func New(
	vs ValidatorSet,
	vrf Verifier,
	d message.Deriver,
) Consensus {
	return Consensus{
		vs:                vs,
		vrf:               vrf,
		deriver:           d,
		currentValidators: make(map[string]struct{}),
	}
}

func (c Consensus) InitSequence(ctx context.Context, sequence uint64) error {
	validators, err := c.vs.GetValidators(ctx, sequence)
	if err != nil {
		return err
	}

	clear(c.currentValidators)
	for _, validator := range validators {
		c.currentValidators[string(validator)] = struct{}{}
	}

	return nil
}
