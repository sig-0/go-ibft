package test

import (
	"context"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/store"
	"github.com/madz-lab/go-ibft/message/types"
	"github.com/madz-lab/go-ibft/sequencer"
)

type IBFTNetwork struct {
	Validators map[string]ibft.Validator
	Messages   *store.Store
	Transport  ibft.Transport
}

func NewIBFTNetwork(validators ...ibft.Validator) IBFTNetwork {
	n := IBFTNetwork{
		Validators: make(map[string]ibft.Validator),
		Messages:   store.New(),
	}

	for _, v := range validators {
		n.Validators[string(v.ID())] = v
	}

	n.Transport = n.GetTransportFn()

	return n
}

func (n IBFTNetwork) ValidatorSet() []ibft.Validator {
	validators := make([]ibft.Validator, 0, len(n.Validators))
	for _, v := range n.Validators {
		validators = append(validators, v)
	}

	sort.Slice(validators, func(i, j int) bool {
		return string(validators[i].ID()) < string(validators[j].ID())
	})

	return validators
}

func (n IBFTNetwork) PoAQuorum() ibft.Quorum {
	validatorCount := len(n.Validators)
	var quorum int

	if (validatorCount-1)/3 == 0 {
		quorum = validatorCount
	} else {
		quorum = int(math.Ceil(2 * float64(validatorCount) / 3))
	}

	q := func(_ uint64, msgs []ibft.Message) bool {
		return len(msgs) >= quorum
	}

	return QuorumFn(q)
}

type TransportOption func(m ibft.Message) bool

func (n IBFTNetwork) FinalizeSequence(sequence uint64, round0Timeout time.Duration) error {
	ch := make(chan *types.FinalizedProposal, len(n.Validators))

	var wg sync.WaitGroup

	wg.Add(len(n.Validators))

	for _, validator := range n.Validators {
		go func(v ibft.Validator) {
			defer wg.Done()

			ctx := ibft.NewIBFTContext(context.Background())
			ctx = ctx.WithFeed(n.Messages.Feed())
			ctx = ctx.WithKeccak(DefaultKeccak)
			ctx = ctx.WithQuorum(n.PoAQuorum())
			ctx = ctx.WithSigRecover(ECRecover)
			ctx = ctx.WithTransport(n.Transport)

			seq := sequencer.New(v, NewIBFTVerifier(n), Round0Timeout)

			ch <- seq.FinalizeSequence(ctx, sequence)
		}(validator)
	}

	wg.Wait()

	var proposals []*types.FinalizedProposal
	for i := 0; i < len(n.Validators); i++ {
		proposals = append(proposals, <-ch)
	}

	// todo: do something with these
	_ = proposals

	return nil

}

func (n IBFTNetwork) GetTransportFn(opts ...TransportOption) TransportFn {
	return func(m ibft.Message) {
		// check opts
		for _, opt := range opts {
			if !opt(m) {
				return
			}
		}

		switch m := m.(type) {
		case *types.MsgProposal:
			n.Messages.ProposalMessages.AddMessage(m)
		case *types.MsgPrepare:
			n.Messages.PrepareMessages.AddMessage(m)
		case *types.MsgCommit:
			n.Messages.CommitMessages.AddMessage(m)
		case *types.MsgRoundChange:
			n.Messages.RoundChangeMessages.AddMessage(m)
		}
	}
}
