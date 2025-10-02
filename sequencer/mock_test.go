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

func (t dummyTransport) MulticastProposal(_ *message.Proposal) {}

func (t dummyTransport) MulticastPrepare(_ *message.Prepare) {}

func (t dummyTransport) MulticastCommit(_ *message.Commit) {}

func (t dummyTransport) MulticastRoundChange(_ *message.RoundChange) {}

type allGoodConsensus struct {
	blockFutureProposal, blockFutureRCC bool
}

func (m allGoodConsensus) AwaitProposal(ctx context.Context, sequence Sequence, store *message.Store) (*message.Proposal, error) {
	messages := store.ProposalMessages.Get(sequence.Number, sequence.Round)
	if len(messages) == 0 {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	return messages[0], nil
}

func (m allGoodConsensus) AwaitFutureProposal(ctx context.Context, sequence Sequence, store *message.Store) (*message.Proposal, error) {
	if m.blockFutureProposal {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	sub, cancel := store.ProposalMessages.Subscribe(sequence.Number, sequence.Round, true)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case fn := <-sub:
			messages := fn()

			if len(messages) == 0 {
				continue
			}

			return messages[0], nil
		}
	}
}

func (m allGoodConsensus) AwaitRoundChange(ctx context.Context, sequence Sequence, store *message.Store) ([]*message.RoundChange, error) {
	messages := store.RoundChangeMessages.Get(sequence.Number, sequence.Round)
	if len(messages) == 0 {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	return messages, nil
}

func (m allGoodConsensus) AwaitFutureRoundChange(ctx context.Context, sequence Sequence, store *message.Store) ([]*message.RoundChange, error) {
	if m.blockFutureRCC {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	sub, cancel := store.RoundChangeMessages.Subscribe(sequence.Number, sequence.Round, true)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case fn := <-sub:
			messages := fn()

			if len(messages) == 0 {
				continue
			}

			return messages, nil
		}
	}
}

func (m allGoodConsensus) AwaitPrepare(ctx context.Context, sequence Sequence, store *message.Store) ([]*message.Prepare, error) {
	messages := store.PrepareMessages.Get(sequence.Number, sequence.Round)
	if len(messages) == 0 {
		<-ctx.Done() // block
		return nil, ctx.Err()
	}

	return messages, nil
}

func (m allGoodConsensus) AwaitCommit(ctx context.Context, sequence Sequence, store *message.Store) ([]*message.Commit, error) {
	messages := store.CommitMessages.Get(sequence.Number, sequence.Round)
	if len(messages) == 0 {
		<-ctx.Done() // block
		return nil, ctx.Err()
	}

	return messages, nil
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
