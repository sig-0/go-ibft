package test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_All_Honest_Validators(t *testing.T) {
	var (
		val1 = HonestValidator(NewIBFTValidator())
		val2 = HonestValidator(NewIBFTValidator())
		val3 = HonestValidator(NewIBFTValidator())
		val4 = HonestValidator(NewIBFTValidator())
	)

	network := NewIBFTNetwork(val1, val2, val3, val4)

	latestSequence, err := network.FinalizeBlocks(10, 101)
	require.NoError(t, err)

	assert.Equal(t, 111, int(latestSequence))
}
