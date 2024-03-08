package consensus

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/madz-lab/go-ibft/sequencer"
	"github.com/madz-lab/go-ibft/test/mock"
)

func Test_EngineConfig(t *testing.T) {
	t.Parallel()

	table := []struct {
		name     string
		cfg      EngineConfig
		expected error
	}{
		{
			name:     "missing validator",
			expected: ErrInvalidConfig,
			cfg: EngineConfig{
				Validator: nil,
			},
		},

		{
			name:     "invalid transport",
			expected: ErrInvalidConfig,
			cfg: EngineConfig{
				Validator: mock.Validator{},
				Transport: sequencer.MessageTransport{},
			},
		},

		{
			name:     "missing quorum",
			expected: ErrInvalidConfig,
			cfg: EngineConfig{
				Validator: mock.Validator{},
				Transport: sequencer.MessageTransport{
					Proposal:    mock.DummyMsgProposalTransport,
					Prepare:     mock.DummyMsgPrepareTransport,
					Commit:      mock.DummyMsgCommitTransport,
					RoundChange: mock.DummyMsgRoundChangeTransport,
				},
				Quorum: nil,
			},
		},

		{
			name:     "missing keccak",
			expected: ErrInvalidConfig,
			cfg: EngineConfig{
				Validator: mock.Validator{},
				Transport: sequencer.MessageTransport{
					Proposal:    mock.DummyMsgProposalTransport,
					Prepare:     mock.DummyMsgPrepareTransport,
					Commit:      mock.DummyMsgCommitTransport,
					RoundChange: mock.DummyMsgRoundChangeTransport,
				},
				Quorum: mock.NonZeroQuorum,
				Keccak: nil,
			},
		},

		{
			name:     "invalid round 0 duration",
			expected: ErrInvalidConfig,
			cfg: EngineConfig{
				Validator: mock.Validator{},
				Transport: sequencer.MessageTransport{
					Proposal:    mock.DummyMsgProposalTransport,
					Prepare:     mock.DummyMsgPrepareTransport,
					Commit:      mock.DummyMsgCommitTransport,
					RoundChange: mock.DummyMsgRoundChangeTransport,
				},
				Quorum:         mock.NonZeroQuorum,
				Keccak:         mock.DummyKeccak,
				Round0Duration: 0,
			},
		},

		{
			name:     "ok",
			expected: nil,
			cfg: EngineConfig{
				Validator: mock.Validator{},
				Transport: sequencer.MessageTransport{
					Proposal:    mock.DummyMsgProposalTransport,
					Prepare:     mock.DummyMsgPrepareTransport,
					Commit:      mock.DummyMsgCommitTransport,
					RoundChange: mock.DummyMsgRoundChangeTransport,
				},
				Quorum:         mock.NonZeroQuorum,
				Keccak:         mock.DummyKeccak,
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
