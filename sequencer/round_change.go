package sequencer

import (
	"bytes"
	"context"

	"github.com/madz-lab/go-ibft/message/types"
)

func (s *Sequencer) getRoundChangeMessages(ctx context.Context, view *types.View, feed MessageFeed) ([]*types.MsgRoundChange, error) {
	sub, cancelSub := feed.SubscribeToRoundChangeMessages(view, false)
	defer cancelSub()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrap := <-sub:
			var validMessages []*types.MsgRoundChange

			for _, msg := range unwrap() {
				if s.isValidRoundChange(view, msg) {
					validMessages = append(validMessages, msg)
				}
			}

			if len(validMessages) == 0 {
				continue
			}

			if !s.quorum.HasQuorumRoundChangeMessages(validMessages...) {
				continue
			}

			return validMessages, nil
		}
	}
}

func (s *Sequencer) isValidRoundChange(view *types.View, msg *types.MsgRoundChange) bool {
	if (msg.View.Sequence != view.Sequence) || (msg.View.Round != view.Round) {
		return false
	}

	if !s.verifier.IsValidator(msg.From, view.Sequence) {
		return false
	}

	var (
		pb = msg.LatestPreparedProposedBlock
		pc = msg.LatestPreparedCertificate
	)

	if pb == nil && pc == nil {
		return true
	}

	if (pb == nil && pc != nil) ||
		(pb != nil && pc == nil) {
		return false
	}

	if !s.isValidPreparedCertificate(view, pc) {
		return false
	}

	proposalHash := s.verifier.Keccak(pb.Bytes())
	if !bytes.Equal(proposalHash, pc.ProposalMessage.ProposalHash) {
		return false
	}

	return true
}

func (s *Sequencer) isValidPreparedCertificate(view *types.View, pc *types.PreparedCertificate) bool {
	if !pc.IsValid(view) {
		return false
	}

	// todo: revisit (+ MsgProposal)
	if !s.quorum.HasQuorumPrepareMessages(pc.PrepareMessages...) {
		return false
	}

	// todo: move to types
	//if !bytes.Equal(pc.ProposalMessage.From, s.verifier.RecoverFrom(
	//	pc.ProposalMessage.Payload(),
	//	pc.ProposalMessage.Signature,
	//)) {
	//	return false
	//}

	if !s.verifier.IsProposer(pc.ProposalMessage.View, pc.ProposalMessage.From) {
		return false
	}

	for _, msg := range pc.PrepareMessages {
		// todo: move to message
		//from := s.verifier.RecoverFrom(msg.Payload(), msg.Signature)
		//if !bytes.Equal(msg.From, from) {
		//	return false
		//}

		if !s.verifier.IsValidator(msg.From, view.Sequence) {
			return false
		}

		if s.verifier.IsProposer(msg.View, msg.From) {
			return false
		}
	}

	return true
}

func (s *Sequencer) multicastRoundChangeMessage(view *types.View) {
	msg := &types.MsgRoundChange{
		View:                        view,
		From:                        s.id,
		LatestPreparedProposedBlock: s.state.latestPreparedProposedBlock,
		LatestPreparedCertificate:   s.state.latestPreparedCertificate,
	}

	sig := s.validator.Sign(msg.Payload())
	msg.Signature = sig

	s.transport.MulticastRoundChange(msg)
}
