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

type MessageOption func(m ibft.Message) bool

func ExcludeMsgIf(opts ...MessageOption) MessageOption {
	return func(m ibft.Message) bool {
		for _, opt := range opts {
			if !opt(m) {
				return true
			}
		}

		return false
	}
}

func IsMsgProposal() MessageOption {
	return func(m ibft.Message) bool {
		_, ok := m.(*types.MsgProposal)
		return ok
	}
}

func IsMsgPrepare() MessageOption {
	return func(m ibft.Message) bool {
		_, ok := m.(*types.MsgPrepare)
		return ok
	}
}

func IsMsgCommit() MessageOption {
	return func(m ibft.Message) bool {
		_, ok := m.(*types.MsgCommit)
		return ok
	}
}

func IsMsgRoundChange() MessageOption {
	return func(m ibft.Message) bool {
		_, ok := m.(*types.MsgRoundChange)
		return ok
	}
}

func HasRound(r uint64) MessageOption {
	return func(m ibft.Message) bool {
		return m.GetRound() == r
	}
}

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

type FinalizedProposals map[string]*types.FinalizedProposal

func (fp FinalizedProposals) From(id []byte) *types.FinalizedProposal {
	return fp[string(id)]
}

func (n IBFTNetwork) FinalizeSequence(sequence uint64, round0Timeout time.Duration) (FinalizedProposals, error) {
	type tuple struct {
		Validator []byte
		Proposal  *types.FinalizedProposal
	}

	var (
		proposals = make(FinalizedProposals)
		ch        = make(chan tuple, len(n.Validators))
		wg        sync.WaitGroup
	)

	defer close(ch)

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

			fp := sequencer.New(v, NewIBFTVerifier(n), round0Timeout).FinalizeSequence(ctx, sequence)

			ch <- tuple{v.ID(), fp}
		}(validator)
	}

	wg.Wait()

	for i := 0; i < len(n.Validators); i++ {
		t := <-ch
		proposals[string(t.Validator)] = t.Proposal
	}

	return proposals, nil
}

func (n IBFTNetwork) WithTransport(opts ...MessageOption) IBFTNetwork {
	n.Transport = n.GetTransportFn(opts...)
	return n
}

func (n IBFTNetwork) GetTransportFn(opts ...MessageOption) TransportFn {
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
