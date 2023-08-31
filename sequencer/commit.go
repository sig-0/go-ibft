package sequencer

import (
	"bytes"
	"context"

	"github.com/madz-lab/go-ibft/message/types"
)

func (s *Sequencer) waitForCommit(ctx context.Context, view *types.View, feed MessageFeed) error {
	sub, cancelSub := feed.SubscribeToCommitMessages(view, false)
	defer cancelSub()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case unwrapCommits := <-sub:
			msgs := unwrapCommits()

			var validCommits []*types.MsgCommit
			for _, msg := range msgs {
				if s.isValidCommit(view, msg) {
					validCommits = append(validCommits, msg)
				}
			}

			if len(validCommits) == 0 {
				continue
			}

			if !s.quorum.HasQuorumCommitMessages(validCommits...) {
				continue
			}

			var commitSeals [][]byte
			for _, msgCommit := range validCommits {
				commitSeals = append(commitSeals, msgCommit.CommitSeal)
			}

			fb := &FinalizedBlock{
				Block:       s.state.acceptedProposal.GetProposedBlock().GetData(),
				Round:       view.Round,
				CommitSeals: commitSeals,
			}

			s.state.finalizedBlock = fb

			return nil
		}
	}
}

func (s *Sequencer) isValidCommit(view *types.View, msg *types.MsgCommit) bool {
	if msg.GetView().GetSequence() != view.GetSequence() || msg.GetView().GetRound() != view.GetRound() {
		return false
	}

	if !bytes.Equal(msg.GetProposalHash(), s.state.acceptedProposal.GetProposalHash()) {
		return false
	}

	if !s.verifier.IsValidator(msg.From, msg.View.Sequence) {
		return false
	}

	// commit_seal = sig(proposal_hash) -> from = recover(proposal_hash, commit_seal)
	if !bytes.Equal(msg.GetFrom(), s.verifier.RecoverFrom(s.state.acceptedProposal.GetProposalHash(), msg.GetCommitSeal())) {
		return false
	}

	return true
}
