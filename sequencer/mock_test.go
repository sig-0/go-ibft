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

func (v mockValidator) BuildProposal(_ uint64) []byte {
	return []byte(v + "_proposal")
}

type dummyTransport struct{}

func (t dummyTransport) MulticastProposal(_ *message.Proposal) {}

func (t dummyTransport) MulticastPrepare(_ *message.Prepare) {}

func (t dummyTransport) MulticastCommit(_ *message.Commit) {}

func (t dummyTransport) MulticastRoundChange(_ *message.RoundChange) {}

type mockConsensus struct {
	awaitProposal    func(ctx context.Context, sequence Sequence, store *message.Store, bool2 bool) (*message.Proposal, error)
	awaitPrepare     func(ctx context.Context, sequence Sequence, store *message.Store) ([]*message.Prepare, error)
	awaitCommit      func(ctx context.Context, sequence Sequence, store *message.Store) ([]*message.Commit, error)
	awaitRoundChange func(ctx context.Context, sequence Sequence, store *message.Store, bool2 bool) ([]*message.RoundChange, error)
}

func (m mockConsensus) AwaitProposal(ctx context.Context, sequence Sequence, store *message.Store, fromHigherRounds bool) (*message.Proposal, error) {
	return m.awaitProposal(ctx, sequence, store, fromHigherRounds)
}

func (m mockConsensus) AwaitRoundChange(ctx context.Context, sequence Sequence, store *message.Store, fromHigherRounds bool) ([]*message.RoundChange, error) {
	return m.awaitRoundChange(ctx, sequence, store, fromHigherRounds)
}

func (m mockConsensus) AwaitPrepare(ctx context.Context, sequence Sequence, store *message.Store) ([]*message.Prepare, error) {
	return m.awaitPrepare(ctx, sequence, store)
}

func (m mockConsensus) AwaitCommit(ctx context.Context, sequence Sequence, store *message.Store) ([]*message.Commit, error) {
	return m.awaitCommit(ctx, sequence, store)
}

type allGoodConsensus struct {
	blockHigherProposal, blockHigherRCC bool
}

func (m allGoodConsensus) AwaitProposal(ctx context.Context, sequence Sequence, store *message.Store, fromHigherRounds bool) (*message.Proposal, error) {
	sub, cancel := store.ProposalMessages.Subscribe(sequence.sequence, sequence.round, fromHigherRounds)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case fn := <-sub:
			messages := fn()
			if len(messages) == 0 || m.blockHigherProposal && fromHigherRounds == true {
				continue
			}

			return messages[0], nil
		}
	}
}

func (m allGoodConsensus) AwaitPrepare(ctx context.Context, sequence Sequence, store *message.Store) ([]*message.Prepare, error) {
	messages := store.PrepareMessages.Get(sequence.sequence, sequence.round)
	if len(messages) == 0 {
		<-ctx.Done() // block
		return nil, ctx.Err()
	}

	return messages, nil
}

func (m allGoodConsensus) AwaitCommit(ctx context.Context, sequence Sequence, store *message.Store) ([]*message.Commit, error) {
	messages := store.CommitMessages.Get(sequence.sequence, sequence.round)
	if len(messages) == 0 {
		<-ctx.Done() // block
		return nil, ctx.Err()
	}

	return messages, nil
}

func (m allGoodConsensus) AwaitRoundChange(ctx context.Context, sequence Sequence, store *message.Store, fromHigherRounds bool) ([]*message.RoundChange, error) {
	sub, cancel := store.RoundChangeMessages.Subscribe(sequence.sequence, sequence.round, fromHigherRounds)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case fn := <-sub:
			messages := fn()

			if len(messages) == 0 || m.blockHigherRCC && fromHigherRounds == true {
				continue
			}

			return messages, nil
		}
	}
}

type mockProposerAlgo func(ctx context.Context, sequence, round uint64) ([]byte, error)

func (m mockProposerAlgo) GetProposer(ctx context.Context, sequence, round uint64) ([]byte, error) {
	return m(ctx, sequence, round)
}
