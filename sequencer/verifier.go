package sequencer

import (
	"context"

	"github.com/sig-0/go-ibft/message"
)

type Verifier interface {
	CheckProposal(ctx context.Context, sequence Sequence, messages []*message.Proposal) (*message.Proposal, error)
	CheckPrepare(ctx context.Context, sequence Sequence, messages []*message.Prepare) ([]*message.Prepare, error)
	CheckCommit(ctx context.Context, sequence Sequence, messages []*message.Commit) ([]*message.Commit, error)
	CheckRoundChange(ctx context.Context, sequence Sequence, messages []*message.RoundChange) ([]*message.RoundChange, error)
}
