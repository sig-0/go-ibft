package sequencer

import (
	"bytes"
	"context"

	"github.com/madz-lab/go-ibft/message/types"
)

func (s *Sequencer) propose(ctx context.Context, view *types.View, feed MessageFeed) error {
	if view.Round == 0 {
		//	build fresh block
		block := s.validator.BuildBlock()
		pb := &types.ProposedBlock{
			Data:  block,
			Round: 0,
		}

		msg := &types.MsgProposal{
			View:          view,
			From:          s.id,
			ProposedBlock: pb,
			ProposalHash:  s.verifier.Keccak(pb.Bytes()),
		}

		sig := s.validator.Sign(msg.Payload())

		msg.Signature = sig

		s.state.acceptedProposal = msg

		s.transport.MulticastProposal(msg)

		return nil
	}

	// todo: higher rounds

	rcc, err := s.getRoundChangeCertificate(ctx, view, feed)
	_, _ = rcc, err

	return nil
}

func (s *Sequencer) waitForProposal(ctx context.Context, view *types.View, feed MessageFeed) error {
	if s.state.acceptedProposal != nil {
		// this node is the proposer
		return nil
	}

	sub, cancelSub := feed.SubscribeToProposalMessages(view, false)
	defer cancelSub()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case unwrapProposals := <-sub:
			messages := unwrapProposals()

			var validProposals []*types.MsgProposal
			for _, msg := range messages {
				if s.isValidProposal(view, msg) {
					validProposals = append(validProposals, msg)
				}
			}

			if len(validProposals) == 0 {
				continue
			}

			if len(validProposals) > 1 {
				//	todo proposer misbehavior
			}

			proposal := validProposals[0]

			s.state.acceptedProposal = proposal

			s.multicastPrepare(view)

			return nil
		}
	}
}

func (s *Sequencer) isValidProposal(view *types.View, msg *types.MsgProposal) bool {
	if msg.View.Sequence != view.Sequence || msg.View.Round != view.Round {
		return false
	}

	if msg.ProposedBlock.Round != view.Round {
		return false
	}

	if s.verifier.IsProposer(view, s.id) {
		return false
	}

	if !s.verifier.IsProposer(view, msg.From) {
		return false
	}

	if !s.verifier.IsValidBlock(msg.ProposedBlock.Data) {
		return false
	}

	if !bytes.Equal(msg.ProposalHash, s.verifier.Keccak(msg.ProposedBlock.Bytes())) {
		return false
	}

	if view.Round == 0 {
		// all checks for round 0 proposal satisfied
		return true
	}

	// todo higher rounds

	return true
}

func (s *Sequencer) multicastPrepare(view *types.View) {
	msg := &types.MsgPrepare{
		View:         view,
		From:         s.validator.ID(),
		ProposalHash: s.state.acceptedProposal.GetProposalHash(),
	}

	msg.Signature = s.validator.Sign(msg.Payload())

	s.transport.MulticastPrepare(msg)
}
