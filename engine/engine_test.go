package engine_test

import (
	"bytes"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
	"github.com/madz-lab/go-ibft/test"
	"github.com/madz-lab/go-ibft/test/mock"

	. "github.com/madz-lab/go-ibft/engine"
)

func Test_EngineConfig(t *testing.T) {
	t.Parallel()

	table := []struct {
		name      string
		validator ibft.Validator
		cfg       Config
		expected  error
	}{
		{
			name:     "missing proposal transport",
			expected: ErrInvalidConfig,
			cfg:      Config{},
		},

		{
			name:     "missing prepare transport",
			expected: ErrInvalidConfig,
			cfg: Config{
				TransportMsgProposal: mock.DummyMsgProposalTransport,
			},
		},

		{
			name:     "missing commit transport",
			expected: ErrInvalidConfig,
			cfg: Config{
				TransportMsgProposal: mock.DummyMsgProposalTransport,
				TransportMsgPrepare:  mock.DummyMsgPrepareTransport,
			},
		},

		{
			name:     "missing round change transport",
			expected: ErrInvalidConfig,
			cfg: Config{
				TransportMsgProposal: mock.DummyMsgProposalTransport,
				TransportMsgPrepare:  mock.DummyMsgPrepareTransport,
				TransportMsgCommit:   mock.DummyMsgCommitTransport,
			},
		},

		{
			name:     "missing quorum",
			expected: ErrInvalidConfig,
			cfg: Config{
				TransportMsgProposal:    mock.DummyMsgProposalTransport,
				TransportMsgPrepare:     mock.DummyMsgPrepareTransport,
				TransportMsgCommit:      mock.DummyMsgCommitTransport,
				TransportMsgRoundChange: mock.DummyMsgRoundChangeTransport,
			},
		},

		{
			name:     "missing keccak",
			expected: ErrInvalidConfig,
			cfg: Config{
				TransportMsgProposal:    mock.DummyMsgProposalTransport,
				TransportMsgPrepare:     mock.DummyMsgPrepareTransport,
				TransportMsgCommit:      mock.DummyMsgCommitTransport,
				TransportMsgRoundChange: mock.DummyMsgRoundChangeTransport,
				Quorum:                  mock.NonZeroQuorum,
			},
		},

		{
			name:     "invalid round 0 duration",
			expected: ErrInvalidConfig,
			cfg: Config{
				TransportMsgProposal:    mock.DummyMsgProposalTransport,
				TransportMsgPrepare:     mock.DummyMsgPrepareTransport,
				TransportMsgCommit:      mock.DummyMsgCommitTransport,
				TransportMsgRoundChange: mock.DummyMsgRoundChangeTransport,
				Quorum:                  mock.NonZeroQuorum,
				Keccak:                  mock.DummyKeccak,
			},
		},

		{
			name:     "ok",
			expected: nil,
			cfg: Config{
				TransportMsgProposal:    mock.DummyMsgProposalTransport,
				TransportMsgPrepare:     mock.DummyMsgPrepareTransport,
				TransportMsgCommit:      mock.DummyMsgCommitTransport,
				TransportMsgRoundChange: mock.DummyMsgRoundChangeTransport,
				Quorum:                  mock.NonZeroQuorum,
				Keccak:                  mock.DummyKeccak,
				Round0Duration:          1 * time.Second,
			},
		},
	}

	for _, tt := range table {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.ErrorIs(t, tt.cfg.IsValid(), tt.expected)
		})
	}
}

func Test_Engine_Add_Message(t *testing.T) {
	t.Parallel()

	var (
		keccak        = mock.NewDummyKeccak("block hash")
		goodSignature = mock.Verifier{IsValidSignatureFn: func(_, _, _ []byte) bool { return true }}
		badSignature  = mock.Verifier{IsValidSignatureFn: func(_, _, _ []byte) bool { return false }}
	)

	table := []struct {
		name      string
		validator ibft.Validator
		cfg       Config
		msg       types.Message
		expected  error
	}{
		{
			name: "invalid message",

			expected: ErrInvalidMessage,
			msg:      &types.MsgPrepare{View: nil},
		},

		{
			name:      "invalid signature",
			validator: mock.Validator{Verifier: badSignature},
			cfg:       Config{Keccak: keccak},

			expected: ErrInvalidMessage,
			msg: &types.MsgPrepare{
				View:      &types.View{Sequence: 101, Round: 1},
				From:      []byte("someone"),
				Signature: []byte("signature"),
				BlockHash: []byte("block hash"),
			},
		},

		{
			name:      "added MsgPrepare",
			validator: mock.Validator{Verifier: goodSignature},
			cfg:       Config{Keccak: keccak},

			expected: nil,
			msg: &types.MsgPrepare{
				View:      &types.View{Sequence: 101, Round: 1},
				From:      []byte("someone"),
				Signature: []byte("signature"),
				BlockHash: []byte("block hash"),
			},
		},

		{
			name:      "added MsgProposal",
			validator: mock.Validator{Verifier: goodSignature},
			cfg:       Config{Keccak: keccak},

			expected: nil,
			msg: &types.MsgProposal{
				View:          &types.View{Sequence: 101, Round: 1},
				From:          []byte("someone"),
				Signature:     []byte("signature"),
				BlockHash:     []byte("block hash"),
				ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 1},
			},
		},

		{
			name:      "added MsgCommit",
			validator: mock.Validator{Verifier: goodSignature},
			cfg:       Config{Keccak: keccak},

			expected: nil,
			msg: &types.MsgCommit{
				View:       &types.View{Sequence: 101, Round: 1},
				From:       []byte("someone"),
				Signature:  []byte("signature"),
				BlockHash:  []byte("block hash"),
				CommitSeal: []byte("commit seal"),
			},
		},

		{
			name:      "added MsgRoundChange",
			validator: mock.Validator{Verifier: goodSignature},
			cfg:       Config{Keccak: keccak},

			expected: nil,
			msg: &types.MsgRoundChange{
				View:      &types.View{Sequence: 101, Round: 1},
				From:      []byte("someone"),
				Signature: []byte("signature"),
			},
		},
	}

	for _, tt := range table {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := NewEngine(tt.validator, tt.cfg)
			assert.ErrorIs(t, e.AddMessage(tt.msg), tt.expected)
		})
	}
}

func Test_Engine_Finalize_Sequence(t *testing.T) {
	t.Parallel()

	var (
		validators     = test.NewValidatorSet("alice", "bob", "chris", "dani")
		keccak         = test.DefaultKeccak()
		quorum         = test.QuorumOf(4)
		network        = test.NewNetwork()
		round0Duration = 1 * time.Second
	)

	engines := make([]Engine, 0, len(validators))
	for _, v := range validators {
		engines = append(engines, NewEngine(v, Config{
			Quorum:                  quorum,
			Keccak:                  keccak,
			TransportMsgProposal:    test.GetTransport[*types.MsgProposal](network),
			TransportMsgPrepare:     test.GetTransport[*types.MsgPrepare](network),
			TransportMsgCommit:      test.GetTransport[*types.MsgCommit](network),
			TransportMsgRoundChange: test.GetTransport[*types.MsgRoundChange](network),
			Round0Duration:          round0Duration,
		}))
	}

	defer close(network)

	go network.Gossip(func(msg types.Message) {
		// all validators receive message
		for _, e := range engines {
			_ = e.AddMessage(msg)
		}
	})

	var (
		results = make(chan SequenceResult, len(validators))
		wg      sync.WaitGroup
	)

	wg.Add(len(validators))

	for _, e := range engines {
		go func(e Engine) {
			defer wg.Done()

			results <- e.FinalizeSequence(context.Background(), 101)
		}(e)
	}

	wg.Wait()
	close(results)

	rr := make([]SequenceResult, 0, len(validators))
	for result := range results {
		rr = append(rr, result)
	}

	block := rr[0].SequenceProposal.Proposal
	round := rr[0].SequenceProposal.Round

	for _, result := range rr[1:] {
		if !bytes.Equal(block, result.SequenceProposal.Proposal) {
			t.FailNow()
		}

		if round != result.SequenceProposal.Round {
			t.FailNow()
		}
	}
}
