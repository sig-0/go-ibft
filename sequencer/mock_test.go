//nolint:dupl //because
package sequencer

import (
	"context"
	"errors"

	"github.com/sig-0/go-ibft/message"
)

var (
	Alice = mockValidator("alice")
	Bob   = mockValidator("bob")
	Chris = mockValidator("chris")
	Nina  = mockValidator("nina")

	DummySignFn = func(_ []byte) []byte { return nil }
)

type dummyTransport struct{}

func (t dummyTransport) MulticastProposal(_ *message.Proposal) {}

func (t dummyTransport) MulticastPrepare(_ *message.Prepare) {}

func (t dummyTransport) MulticastCommit(_ *message.Commit) {}

func (t dummyTransport) MulticastRoundChange(_ *message.RoundChange) {}

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

type allGoodVrf struct {
}

func (m allGoodVrf) CheckProposal(ctx context.Context, sequence Sequence, messages []*message.Proposal) (*message.Proposal, error) {
	if len(messages) == 0 {
		return nil, errors.New("messages is empty")
	}

	return messages[0], nil
}

func (m allGoodVrf) CheckPrepare(ctx context.Context, sequence Sequence, messages []*message.Prepare) ([]*message.Prepare, error) {
	if len(messages) == 0 {
		return nil, errors.New("messages is empty")
	}

	return messages, nil
}

func (m allGoodVrf) CheckCommit(ctx context.Context, sequence Sequence, messages []*message.Commit) ([]*message.Commit, error) {
	if len(messages) == 0 {
		return nil, errors.New("messages is empty")
	}

	return messages, nil
}

func (m allGoodVrf) CheckRoundChange(ctx context.Context, sequence Sequence, messages []*message.RoundChange) ([]*message.RoundChange, error) {
	if len(messages) == 0 {
		return nil, errors.New("messages is empty")
	}

	return messages, nil
}

type mockProposerAlgo func(ctx context.Context, sequence, round uint64) ([]byte, error)

func (m mockProposerAlgo) GetProposer(ctx context.Context, sequence, round uint64) ([]byte, error) {
	return m(ctx, sequence, round)
}
