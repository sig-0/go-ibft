## Overview

`go-ibft` provides a light-weight engine implementation of the IBFT 2.0 block finalization algorithm. 

Due to its simple API and minimal client dependencies (see `ibft.go`) the engine is designed to work in parallel with a block syncing mechanism (as described in the original document). 
For the full protocol specification, see the official doc.

## Installation

Required Go version `1.22`

From your project root directory run:

`$ go get github.com/sig-0/go-ibft`

## Usage Example

```go
package main

import (
	"context"
	"time"

	"github.com/sig-0/go-ibft/message"
	"github.com/sig-0/go-ibft/message/store"
	"github.com/sig-0/go-ibft/sequencer"
)

func main() {
	var (
		v                 sequencer.Validator
		vs                sequencer.ValidatorSet
		transport         sequencer.Transport
		signatureVerifier message.SignatureVerifier
		keccak            sequencer.KeccakFn
		round0Duration    time.Duration
	)

	msgStore := store.NewMsgStore(signatureVerifier)
	
	go func() {
		var msg message.Message

		// ...

		// receive consensus messages from the network
		_ = msgStore.Add(msg)
	}()

	cfg := sequencer.Config{
		Validator:         v,
		ValidatorSet:      vs,
		Transport:         transport,
		Feed:              msgStore.Feed(),
		Keccak:            keccak,
		SignatureVerifier: signatureVerifier,
		Round0Duration:    round0Duration,
	}

	s := sequencer.NewSequencer(cfg)

	// finalize some proposal for give sequence
	_ = s.Finalize(context.Background(), 101)
}

```
