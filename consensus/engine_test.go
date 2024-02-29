package consensus

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Config(t *testing.T) {
	cfg := EngineConfig{
		Validator: nil,
		Verifier:  nil,
		Transport: nil,
		Quorum:    nil,
		Keccak:    nil,
	}

	assert.NoError(t, cfg.IsValid())
}
