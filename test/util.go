package test

import (
	"github.com/madz-lab/go-ibft"
)

type IBFTNetwork struct {
	Validators map[string]ibft.Validator
}

func NewIBFTNetwork(validators ...IBFTValidator) IBFTNetwork {
	n := IBFTNetwork{
		make(map[string]ibft.Validator),
	}

	for _, v := range validators {
		n.Validators[string(v.id)] = v
	}

	return n
}
