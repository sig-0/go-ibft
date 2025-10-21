package consensus

import (
	"bytes"
	"context"

	"github.com/sig-0/go-ibft/message"
	"github.com/sig-0/go-ibft/sequencer"
)

func (c Consensus) AwaitPrepare(
	ctx context.Context,
	seq sequencer.Sequence,
	store *message.Store,
) ([]*message.Prepare, error) {
	seen := make(map[string]struct{})
	valid := make([]*message.Prepare, 0)
	sequence := seq.Number
	round := seq.Round

	sub, cancel := store.PrepareMessages.Subscribe(sequence)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrap := <-sub:
			messages := unwrap()
			for _, msg := range messages {
				if msg.Round != round {
					// only interested in higher rounds
					continue
				}

				if _, ok := seen[string(msg.Signature)]; ok {
					continue
				}

				seen[string(msg.Signature)] = struct{}{}

				if !c.isValidPrepare(ctx, seq, msg) {
					continue
				}

				valid = append(valid, msg)
			}
		}

		getValidators := func(messages ...*message.Prepare) [][]byte {
			validators := make([][]byte, 0, len(messages))
			for _, msg := range messages {
				validators = append(validators, msg.Signature)
			}

			return validators
		}

		ok, err := c.vs.CheckQuorum(ctx, sequence, getValidators(valid...))
		if err != nil {
			// todo: log
			continue
		}

		if !ok {
			continue
		}

		return valid, nil
	}
}

func (c Consensus) isValidPrepare(_ context.Context, seq sequencer.Sequence, msg *message.Prepare) bool {
	// sender must be among validator for current sequence
	if _, ok := c.currentValidators[string(msg.Sender)]; !ok {
		return false
	}

	// block hash must match accepted block hash
	if !bytes.Equal(msg.BlockHash, seq.Proposal.BlockHash) {
		return false
	}

	return true
}
