package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_All_Honest_Validators(t *testing.T) {
	var (
		val1 = HonestValidator(NewIBFTValidator())
		val2 = HonestValidator(NewIBFTValidator())
		val3 = HonestValidator(NewIBFTValidator())
		val4 = HonestValidator(NewIBFTValidator())
	)

	require.NoError(t, NewIBFTNetwork(val1, val2, val3, val4).FinalizeSequence(101, Round0Timeout))
}
