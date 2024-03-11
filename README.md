## Overview

`go-ibft` provides a light-weight engine implementation of the IBFT 2.0 block finalization algorithm. 

Due to its simple API and minimal client dependencies (see `ibft.go`) the engine is designed to work in parallel with a block syncing mechanism (as described in the original document). 
For the full protocol specification, see the official doc.

## Installation

Required Go version `1.22`

From your project root directory run:

`$ go get github.com/madz-lab/go-ibft`

## Usage Examples

```go
package main

import (
	"context"
	"time"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/engine"
	"github.com/madz-lab/go-ibft/message/types"
)

func main() {

	var (
		// Provide external dependencies
		validator      ibft.Validator
		keccak         ibft.Keccak
		quorum         ibft.Quorum
		round0Duration time.Duration

		proposalTransport    ibft.Transport[*types.MsgProposal]
		prepareTransport     ibft.Transport[*types.MsgPrepare]
		commitTransport      ibft.Transport[*types.MsgCommit]
		roundChangeTransport ibft.Transport[*types.MsgRoundChange]
	)
	
	// ...

	cfg := engine.Config{
		TransportMsgProposal:    proposalTransport,
		TransportMsgPrepare:     prepareTransport,
		TransportMsgCommit:      commitTransport,
		TransportMsgRoundChange: roundChangeTransport,
		Quorum:                  quorum,
		Keccak:                  keccak,
		Round0Duration:          round0Duration,
	}

	e := engine.NewEngine(validator, cfg)

	go func() {
		var msg types.Message

		// Receive consensus messages gossiped in the network
		_ = e.AddMessage(msg)
	}()

	// await proposal to be finalized for current sequence
	_ = e.FinalizeSequence(context.Background(), 101)
}

```
