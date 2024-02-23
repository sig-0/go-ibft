package test

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/madz-lab/go-ibft/message/store"
	"github.com/madz-lab/go-ibft/message/types"
	"github.com/madz-lab/go-ibft/sequencer"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/madz-lab/go-ibft"
)

const Round0Timeout = 10 * time.Second

var (
	Keccak ibft.Keccak = KeccakFn(func(data []byte) []byte {
		return crypto.Keccak256(data)
	})

	SigRecover ibft.SigRecover = SigRecoverFn(func(digest, sig []byte) []byte {
		pubKey, err := crypto.SigToPub(digest, sig)
		if err != nil {
			panic(fmt.Errorf("failed to extract pub key: %w", err).Error())
		}

		return crypto.PubkeyToAddress(*pubKey).Bytes()
	})
)

type TransportFn func(ibft.Message)

func (f TransportFn) Multicast(msg ibft.Message) {
	f(msg)
}

type QuorumFn func(uint64, []ibft.Message) bool

func (f QuorumFn) HasQuorum(sequence uint64, messages []ibft.Message) bool {
	return f(sequence, messages)
}

type SigRecoverFn func(digest, sig []byte) []byte

func (f SigRecoverFn) From(digest, sig []byte) []byte {
	return f(digest, sig)
}

type KeccakFn func([]byte) []byte

func (f KeccakFn) Hash(data []byte) []byte {
	return f(data)
}

type IBFTNetwork struct {
	Validators map[string]ibft.Validator
	Messages   *store.Store
}

func NewIBFTNetwork(validators ...ibft.Validator) IBFTNetwork {
	n := IBFTNetwork{
		Validators: make(map[string]ibft.Validator),
		Messages:   store.New(),
	}

	for _, v := range validators {
		n.Validators[string(v.ID())] = v
	}

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

func (n IBFTNetwork) FinalizeBlocks(num int, startSequence uint64) (uint64, error) {
	for i := 0; i < num; i++ {
		if err := n.FinalizeSequence(startSequence); err != nil {
			return 0, err
		}

		startSequence++
	}

	return startSequence, nil
}

func (n IBFTNetwork) FinalizeSequence(sequence uint64) error {
	ch := make(chan *types.FinalizedProposal, len(n.Validators))

	var wg sync.WaitGroup

	wg.Add(len(n.Validators))

	for _, validator := range n.Validators {
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

			vrf = NewIBFTVerifier(n)
			feed = n.Messages.Feed()
			keccak = Keccak
			sigrecover = SigRecover
			quorum = n.PoAQuorum()
			transport = TransportFn(func(m ibft.Message) {
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
			})

			ctx := ibft.NewIBFTContext(context.Background())
			ctx = ctx.WithFeed(feed)
			ctx = ctx.WithKeccak(keccak)
			ctx = ctx.WithQuorum(quorum)
			ctx = ctx.WithSigRecover(sigrecover)
			ctx = ctx.WithTransport(transport)

			seq := sequencer.New(v, vrf, Round0Timeout)

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

type ECDSAKey struct {
	*ecdsa.PrivateKey
}

func NewECDSAKey() ECDSAKey {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic(fmt.Errorf("failed to generate ecdsa key: %w", err).Error())
	}

	return ECDSAKey{key}
}

func (k ECDSAKey) Sign(digest []byte) []byte {
	sig, err := crypto.Sign(digest, k.PrivateKey)
	if err != nil {
		panic(fmt.Errorf("failed to sign: %w", err).Error())
	}

	return sig
}

func (k ECDSAKey) ID() []byte {
	pubKeyECDSA, ok := k.Public().(*ecdsa.PublicKey)
	if !ok {
		panic("failed to cast pub key to *ecdsa.PublicKey")
	}

	return crypto.PubkeyToAddress(*pubKeyECDSA).Bytes()
}
