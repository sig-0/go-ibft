//nolint:dupl //because
package sequencer

import (
	"context"

	"github.com/sig-0/go-ibft/message"
)

var (
	Alice = mockValidator("alice")
	Bob   = mockValidator("bob")
	Chris = mockValidator("chris")
	Nina  = mockValidator("nina")
)

type mockValidator string

func (v mockValidator) Address() []byte {
	return []byte(v)
}

func (v mockValidator) Sign(_ []byte) []byte {
	return []byte(v + "_sig")
}

func (v mockValidator) BuildProposal(ctx context.Context, sequence uint64) ([]byte, error) {
	return []byte(v + "_proposal"), nil
}

type dummyTransport struct{}

func (t dummyTransport) MulticastProposal(ctx context.Context, msg *message.Proposal) {}

func (t dummyTransport) MulticastPrepare(ctx context.Context, msg *message.Prepare) {}

func (t dummyTransport) MulticastCommit(ctx context.Context, msg *message.Commit) {}

func (t dummyTransport) MulticastRoundChange(ctx context.Context, msg *message.RoundChange) {}

type allGoodConsensus struct {
	blockFutureProposal, blockFutureRCC bool
}

func (m allGoodConsensus) AwaitProposal(ctx context.Context, sequence Sequence, store *message.Store) (*message.Proposal, error) {
	messages := store.ProposalMessages.GetSequence(sequence.Number)

	filtered := make([]*message.Proposal, 0, len(messages))
	for _, msg := range messages {
		if msg.Round != sequence.Round {
			continue
		}

		filtered = append(filtered, msg)
	}

	if len(filtered) == 0 {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	return filtered[0], nil
}

func (m allGoodConsensus) AwaitFutureProposal(ctx context.Context, sequence Sequence, store *message.Store) (*message.Proposal, error) {
	if m.blockFutureProposal {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	sub, cancel := store.ProposalMessages.Subscribe(sequence.Number)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case fn := <-sub:
			messages := fn()

			filtered := make([]*message.Proposal, 0, len(messages))
			for _, msg := range messages {
				if msg.Round <= sequence.Round {
					continue
				}

				filtered = append(filtered, msg)
			}

			if len(filtered) == 0 {
				continue
			}

			return filtered[0], nil
		}
	}
}

func (m allGoodConsensus) AwaitRoundChange(ctx context.Context, sequence Sequence, store *message.Store) ([]*message.RoundChange, error) {
	messages := store.RoundChangeMessages.GetSequence(sequence.Number)
	filtered := make([]*message.RoundChange, 0, len(messages))
	for _, msg := range messages {
		if msg.Round != sequence.Round {
			continue
		}

		filtered = append(filtered, msg)
	}

	if len(filtered) == 0 {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	return filtered, nil
}

func (m allGoodConsensus) AwaitFutureRoundChange(ctx context.Context, sequence Sequence, store *message.Store) ([]*message.RoundChange, error) {
	if m.blockFutureRCC {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	sub, cancel := store.RoundChangeMessages.Subscribe(sequence.Number)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case fn := <-sub:
			messages := fn()

			filtered := make([]*message.RoundChange, 0, len(messages))
			for _, msg := range messages {
				if msg.Round <= sequence.Round {
					continue
				}

				filtered = append(filtered, msg)
			}

			if len(filtered) == 0 {
				continue
			}

			return filtered, nil
		}
	}
}

func (m allGoodConsensus) AwaitPrepare(ctx context.Context, sequence Sequence, store *message.Store) ([]*message.Prepare, error) {
	messages := store.PrepareMessages.GetSequence(sequence.Number)
	filtered := make([]*message.Prepare, 0, len(messages))
	for _, msg := range messages {
		if msg.Round != sequence.Round {
			continue
		}

		filtered = append(filtered, msg)
	}

	if len(filtered) == 0 {
		<-ctx.Done() // block
		return nil, ctx.Err()
	}

	return filtered, nil
}

func (m allGoodConsensus) AwaitCommit(ctx context.Context, sequence Sequence, store *message.Store) ([]*message.Commit, error) {
	messages := store.CommitMessages.GetSequence(sequence.Number)
	filtered := make([]*message.Commit, 0, len(messages))
	for _, msg := range messages {
		if msg.Round != sequence.Round {
			continue
		}

		filtered = append(filtered, msg)
	}

	if len(filtered) == 0 {
		<-ctx.Done() // block
		return nil, ctx.Err()
	}

	return filtered, nil
}

type consensusOfTwo struct {
	allGoodConsensus
}

func (m consensusOfTwo) AwaitCommit(ctx context.Context, sequence Sequence, store *message.Store) ([]*message.Commit, error) {
	messages, _ := m.allGoodConsensus.AwaitCommit(ctx, sequence, store)
	if len(messages) < 2 {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	return messages, nil
}

type mockProposerAlgo func(ctx context.Context, sequence, round uint64) ([]byte, error)

func (m mockProposerAlgo) GetProposer(ctx context.Context, sequence, round uint64) ([]byte, error) {
	return m(ctx, sequence, round)
}
