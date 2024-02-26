package test

import (
	"bytes"
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

	q := func(msgs []ibft.Message) bool {
		return len(msgs) >= quorum
	}

	return QuorumFn(q)
}

func (n IBFTNetwork) FinalizeSequence(sequence uint64, round0Timeout time.Duration) ([]*types.FinalizedProposal, error) {
	var (
		proposals    = make([]*types.FinalizedProposal, 0)
		wg           sync.WaitGroup
		proposalsMux sync.Mutex
	)

	wg.Add(len(n.Validators))

	for _, validator := range n.Validators {
		go func(v ibft.Validator) {
			defer wg.Done()

			ctx := ibft.NewIBFTContext(context.Background())
			ctx = ctx.WithFeed(n.Messages.Feed())
			ctx = ctx.WithKeccak(DefaultKeccak)
			ctx = ctx.WithQuorum(n.PoAQuorum())
			ctx = ctx.WithTransport(n.Transport)

			fp := sequencer.New(v, NewIBFTVerifier(n), round0Timeout).FinalizeSequence(ctx, sequence)

			proposalsMux.Lock()
			defer proposalsMux.Unlock()

			proposals = append(proposals, fp)

		}(validator)
	}

	wg.Wait()

	return proposals, nil
}

func (n IBFTNetwork) WithTransport(opts ...MessageOption) IBFTNetwork {
	n.Transport = n.GetTransportFn(opts...)
	return n
}

func (n IBFTNetwork) GetTransportFn(opts ...MessageOption) TransportFn {
	return func(m ibft.Message) {
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

func AllValidProposals(network IBFTNetwork, proposals []*types.FinalizedProposal) bool {
	if len(network.Validators) != len(proposals) {
		return false // all validators must have finalized proposals
	}

	allProposals := make([][]byte, 0, len(proposals))
	for _, p := range proposals {
		allProposals = append(allProposals, p.Proposal)
	}

	p := allProposals[0]
	for _, pp := range allProposals[1:] {
		if !bytes.Equal(p, pp) {
			return false
		}
	}

	allRounds := make([]uint64, 0, len(proposals))
	for _, p := range proposals {
		allRounds = append(allRounds, p.Round)
	}

	r := allRounds[0]
	for _, round := range allRounds[1:] {
		if r != round {
			return false
		}
	}

	seals := make(map[string][]byte)
	for _, p := range proposals {
		for _, seal := range p.Seals {
			seals[string(seal.From)] = seal.CommitSeal
		}
	}

	if len(proposals) != len(seals) {
		return false
	}

	return true
}
