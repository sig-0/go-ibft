package test

import (
	"context"
	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
	"github.com/madz-lab/go-ibft/sequencer"
	"sync"
	"testing"
	"time"
)

func Test_All_Honest_Validators(t *testing.T) {
	validators := []IBFTValidator{
		{[]byte("validator 1")},
		{[]byte("validator 2")},
		{[]byte("validator 3")},
	}

	network := NewIBFTNetwork(validators...)

	var wg sync.WaitGroup

	wg.Add(4)
	ch := make(chan *types.FinalizedProposal, 4)

	for _, validator := range network.Validators {
		go func(v ibft.Validator) {
			defer wg.Done()

			var (
				vrf        ibft.Verifier
				feed       ibft.MessageFeed
				keccak     ibft.Keccak
				sigrecover ibft.SigRecover
				quorum     ibft.Quorum
				transport  ibft.Transport
			)

			ctx := ibft.NewIBFTContext(context.Background())
			ctx = ctx.WithFeed(feed)
			ctx = ctx.WithKeccak(keccak)
			ctx = ctx.WithQuorum(quorum)
			ctx = ctx.WithSigRecover(sigrecover)
			ctx = ctx.WithTransport(transport)

			seq := sequencer.New(v, nil, 10*time.Second)

			ch <- seq.FinalizeSequence(ctx, 101)
		}(validator)
	}

	wg.Wait()

	var proposals []*types.FinalizedProposal
	for i := 0; i < 4; i++ {
		proposals = append(proposals, <-ch)
	}

}
