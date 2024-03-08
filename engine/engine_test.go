package consensus

import (
	"github.com/madz-lab/go-ibft/sequencer"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Config(t *testing.T) {
	t.Skip()

	cfg := EngineConfig{
		Validator: nil,
		Transport: sequencer.MessageTransport{},
		Quorum:    nil,
		Keccak:    nil,
	}

	assert.NoError(t, cfg.IsValid())
}
