package consensus

import (
	"bytes"
	"context"
	"sync"

	"github.com/sig-0/go-ibft/message"
	"github.com/sig-0/go-ibft/sequencer"
)

func (c Consensus) AwaitPrepare(
	ctx context.Context,
	seq sequencer.Sequence,
	store *message.Store,
) ([]*message.Prepare, error) {
	var (
		sequence = seq.Number
		round    = seq.Round
		seen     = make(map[string]struct{})
		valid    = make([]*message.Prepare, 0)
	)

	sub, cancel := store.PrepareMessages.Subscribe(sequence)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrap := <-sub:
			messages := unwrap()
			candidates := make([]*message.Prepare, 0, len(messages))
			for _, msg := range messages {
				if msg.Round != round {
					continue
				}

				if _, ok := seen[string(msg.Signature)]; ok {
					continue
				}

				seen[string(msg.Signature)] = struct{}{}

				candidates = append(candidates, msg)
			}

			valid = filterValidPrepareMessages(ctx, c, candidates, seq)
			validators := make([][]byte, 0, len(valid))
			for _, msg := range valid {
				validators = append(validators, msg.Sender)
			}

			ok, err := c.vs.CheckQuorum(ctx, sequence, validators)
			if err != nil || !ok {
				continue
			}

			return valid, nil
		}
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

func filterValidPrepareMessages(
	ctx context.Context,
	cons Consensus,
	messages []*message.Prepare,
	sequence sequencer.Sequence,
) []*message.Prepare {
	filtered := make([]*message.Prepare, 0, len(messages))

	var (
		mux sync.Mutex
		wg  sync.WaitGroup
	)

	for _, msg := range messages {
		wg.Add(1)

		go func(msg *message.Prepare) {
			defer wg.Done()

			if !cons.isValidPrepare(ctx, sequence, msg) {
				return
			}

			mux.Lock()
			defer mux.Unlock()

			filtered = append(filtered, msg)
		}(msg)
	}

	wg.Wait()

	return filtered
}

func filterValidCommitMessages(
	ctx context.Context,
	cons Consensus,
	messages []*message.Commit,
	sequence sequencer.Sequence,
) []*message.Commit {
	filtered := make([]*message.Commit, 0, len(messages))

	var (
		mux sync.Mutex
		wg  sync.WaitGroup
	)

	for _, msg := range messages {
		wg.Add(1)

		go func(msg *message.Commit) {
			defer wg.Done()

			if !cons.isValidCommit(ctx, sequence, msg) {
				return
			}

			mux.Lock()
			defer mux.Unlock()

			filtered = append(filtered, msg)
		}(msg)
	}

	wg.Wait()

	return filtered
}
