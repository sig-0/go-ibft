package consensus

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/madz-lab/go-ibft/message/types"
	"github.com/madz-lab/go-ibft/sequencer"
	"github.com/madz-lab/go-ibft/test/mock"
)

func Test_EngineConfig(t *testing.T) {
	t.Parallel()

	var (
		validator = mock.Validator{}
		quorum    = mock.NonZeroQuorum
		keccak    = mock.DummyKeccak
		transport = sequencer.MessageTransport{
			Proposal:    mock.DummyMsgProposalTransport,
			Prepare:     mock.DummyMsgPrepareTransport,
			Commit:      mock.DummyMsgCommitTransport,
			RoundChange: mock.DummyMsgRoundChangeTransport,
		}
	)

	table := []struct {
		name     string
		cfg      EngineConfig
		expected error
	}{
		{
			name:     "missing validator",
			expected: ErrInvalidConfig,
			cfg:      EngineConfig{Validator: nil},
		},

		{
			name:     "invalid transport",
			expected: ErrInvalidConfig,
			cfg: EngineConfig{
				Validator: validator,
				Transport: sequencer.MessageTransport{},
			},
		},

		{
			name:     "missing quorum",
			expected: ErrInvalidConfig,
			cfg: EngineConfig{
				Validator: validator,
				Transport: transport,
				Quorum:    nil,
			},
		},

		{
			name:     "missing keccak",
			expected: ErrInvalidConfig,
			cfg: EngineConfig{
				Validator: validator,
				Transport: transport,
				Quorum:    quorum,
				Keccak:    nil,
			},
		},

		{
			name:     "invalid round 0 duration",
			expected: ErrInvalidConfig,
			cfg: EngineConfig{
				Validator:      validator,
				Transport:      transport,
				Quorum:         quorum,
				Keccak:         keccak,
				Round0Duration: 0,
			},
		},

		{
			name:     "ok",
			expected: nil,
			cfg: EngineConfig{
				Validator:      validator,
				Transport:      transport,
				Quorum:         quorum,
				Keccak:         keccak,
				Round0Duration: 1 * time.Second,
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
		name     string
		cfg      EngineConfig
		msg      types.Message
		expected error
	}{
		{
			name:     "invalid message",
			expected: ErrInvalidMessage,
			msg:      &types.MsgPrepare{View: nil},
		},

		{
			name:     "invalid signature",
			expected: ErrInvalidMessage,

			cfg: EngineConfig{
				Keccak:    keccak,
				Validator: mock.Validator{Verifier: badSignature},
			},
			msg: &types.MsgPrepare{
				View:      &types.View{Sequence: 101, Round: 1},
				From:      []byte("someone"),
				Signature: []byte("signature"),
				BlockHash: []byte("block hash"),
			},
		},

		{
			name:     "added MsgPrepare",
			expected: nil,

			cfg: EngineConfig{
				Keccak:    keccak,
				Validator: mock.Validator{Verifier: goodSignature},
			},
			msg: &types.MsgPrepare{
				View:      &types.View{Sequence: 101, Round: 1},
				From:      []byte("someone"),
				Signature: []byte("signature"),
				BlockHash: []byte("block hash"),
			},
		},

		{
			name:     "added MsgProposal",
			expected: nil,

			cfg: EngineConfig{
				Keccak:    keccak,
				Validator: mock.Validator{Verifier: goodSignature},
			},
			msg: &types.MsgProposal{
				View:          &types.View{Sequence: 101, Round: 1},
				From:          []byte("someone"),
				Signature:     []byte("signature"),
				BlockHash:     []byte("block hash"),
				ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 1},
			},
		},

		{
			name:     "added MsgCommit",
			expected: nil,

			cfg: EngineConfig{
				Keccak:    keccak,
				Validator: mock.Validator{Verifier: goodSignature},
			},
			msg: &types.MsgCommit{
				View:       &types.View{Sequence: 101, Round: 1},
				From:       []byte("someone"),
				Signature:  []byte("signature"),
				BlockHash:  []byte("block hash"),
				CommitSeal: []byte("commit seal"),
			},
		},

		{
			name:     "added MsgRoundChange",
			expected: nil,

			cfg: EngineConfig{
				Keccak:    keccak,
				Validator: mock.Validator{Verifier: goodSignature},
			},
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

			e := NewEngine(tt.cfg)
			assert.ErrorIs(t, e.AddMessage(tt.msg), tt.expected)
		})
	}
}
