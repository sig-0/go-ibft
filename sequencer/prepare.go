package sequencer

import (
	"bytes"
	"context"

	"github.com/madz-lab/go-ibft/message/types"
)

func (s *Sequencer) waitForPrepare(ctx context.Context, view *types.View, feed MessageFeed) error {
	sub, cancelSub := feed.SubscribeToPrepareMessages(view, false)
	defer cancelSub()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case unwrapPrepares := <-sub:
			msgs := unwrapPrepares()

			var validPrepares []*types.MsgPrepare
			for _, msg := range msgs {
				if s.isValidPrepare(view, msg) {
					validPrepares = append(validPrepares, msg)
				}
			}

			if !s.quorum.HasQuorumPrepareMessages(validPrepares...) {
				continue
			}

			pc := &types.PreparedCertificate{
				ProposalMessage: s.state.acceptedProposal,
				PrepareMessages: validPrepares,
			}

			pb := s.state.acceptedProposal.GetProposedBlock()

			s.state.latestPreparedCertificate = pc
			s.state.latestPreparedProposedBlock = pb

			s.multicastCommit(view)

			return nil
		}
	}
}

func (s *Sequencer) isValidPrepare(view *types.View, msg *types.MsgPrepare) bool {
	if msg.View.Sequence != view.Sequence || msg.View.Round != view.Round {
		return false
	}

	if !bytes.Equal(msg.ProposalHash, s.state.acceptedProposal.ProposalHash) {
		return false
	}

	return true
}

func (s *Sequencer) multicastCommit(view *types.View) {
	msg := &types.MsgCommit{
		View:         view,
		From:         s.validator.ID(),
		ProposalHash: s.state.acceptedProposal.GetProposalHash(),
	}

	pb := s.state.acceptedProposal.GetProposedBlock()
	cs := s.validator.Sign(s.verifier.Keccak(pb.Bytes()))

	msg.CommitSeal = cs

	sig := s.validator.Sign(msg.Payload())

	msg.Signature = sig

	s.transport.MulticastCommit(msg)
}
