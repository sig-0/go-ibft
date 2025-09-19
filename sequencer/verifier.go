package sequencer

import (
	"context"

	"github.com/sig-0/go-ibft/message"
)

type Verifier interface {
	// IsValidator checks if id is part of consensus for given sequence
	IsValidator(addr []byte, sequence uint64) bool

	// IsProposer asserts if id is the elected proposer for given sequence and round
	IsProposer(addr []byte, sequence, round uint64) bool

	// HasQuorum returns true if messages accumulate consensus for a particular sequence
	HasQuorum(addresses [][]byte, sequence uint64) bool

	// IsValidProposal checks if the provided proposal is valid for given sequence
	IsValidProposal(proposal []byte, sequence uint64) bool

	// IsValidSignature checks if the signature came from signer over some digest
	IsValidSignature(signer, digest, signature []byte) error
}

type Vrf interface {
	CheckProposal(ctx context.Context, sequence *Sequence, messages []*message.Proposal) (*message.Proposal, error)
	CheckPrepare(ctx context.Context, sequence *Sequence, messages []*message.Prepare) ([]*message.Prepare, error)
	CheckCommit(ctx context.Context, sequence *Sequence, messages []*message.Commit) ([]*message.Commit, error)
	CheckRoundChange(ctx context.Context, sequence *Sequence, messages []*message.RoundChange) ([]*message.RoundChange, error)
}
