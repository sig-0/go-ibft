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

	rcc := s.state.roundChangeCertificate
	if rcc == nil {
		msgs, err := s.getRoundChangeMessages(ctx, view, feed)
		if err != nil {
			return err
		}

		rcc = &types.RoundChangeCertificate{Messages: msgs}
	}

	roundsAndPreparedBlocks := make(map[uint64][]byte)

	for _, msg := range rcc.Messages {
		pb := msg.LatestPreparedProposedBlock
		pc := msg.LatestPreparedCertificate

		if pb == nil || pc == nil {
			continue
		}

		roundsAndPreparedBlocks[pc.ProposalMessage.View.Round] = pb.Data
	}

	var (
		maxRound      uint64
		maxRoundBlock []byte
	)

	for round, block := range roundsAndPreparedBlocks {
		if round >= maxRound {
			maxRound = round
			maxRoundBlock = block
		}
	}

	if maxRoundBlock == nil {
		// no block found in rcc, build new
		newBlock := s.validator.BuildBlock()

		pb := &types.ProposedBlock{
			Data:  newBlock,
			Round: view.Round,
		}

		msg := &types.MsgProposal{
			View:                   view,
			From:                   s.id,
			ProposedBlock:          pb,
			ProposalHash:           s.verifier.Keccak(pb.Bytes()),
			RoundChangeCertificate: rcc,
		}

		sig := s.validator.Sign(msg.Payload())

		msg.Signature = sig

		s.state.acceptedProposal = msg

		s.transport.MulticastProposal(msg)

		return nil

	}

	pb := &types.ProposedBlock{
		Data:  maxRoundBlock,
		Round: view.Round,
	}

	msg := &types.MsgProposal{
		View:                   view,
		From:                   s.id,
		ProposedBlock:          pb,
		ProposalHash:           s.verifier.Keccak(pb.Bytes()),
		RoundChangeCertificate: rcc,
	}

	sig := s.validator.Sign(msg.Payload())
	msg.Signature = sig

	s.state.acceptedProposal = msg

	s.transport.MulticastProposal(msg)

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
				//	todo proposer misbehaviorx
			}

			s.state.acceptedProposal = validProposals[0]

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

	if !s.verifier.IsValidator(msg.From, view.Sequence) {
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

	// verify rcc
	rcc := msg.RoundChangeCertificate
	if rcc == nil {
		return false
	}

	if !rcc.IsValid(view) {
		return false
	}

	if !s.quorum.HasQuorumRoundChangeMessages(rcc.Messages...) {
		return false
	}

	// todo: extract
	for _, msg := range rcc.Messages {
		// todo: move to types
		//if !bytes.Equal(msg.From, s.verifier.RecoverFrom(msg.Payload(), msg.Signature)) {
		//	return false
		//}

		if !s.verifier.IsValidator(msg.From, view.Sequence) {
			return false
		}
	}

	roundsAndPreparedProposalHashes := make(map[uint64][]byte)
	for _, msg := range rcc.Messages {
		pc := msg.LatestPreparedCertificate
		if pc == nil {
			continue
		}

		if !s.isValidPreparedCertificate(view, pc) {
			continue
		}

		roundsAndPreparedProposalHashes[pc.ProposalMessage.View.Round] = pc.ProposalMessage.ProposalHash
	}

	if len(roundsAndPreparedProposalHashes) == 0 {
		return true
	}

	var (
		maxRound             uint64
		maxRoundProposalHash []byte
	)

	for round, proposalHash := range roundsAndPreparedProposalHashes {
		if round >= maxRound {
			maxRound = round
			maxRoundProposalHash = proposalHash
		}
	}

	pb := &types.ProposedBlock{
		Data:  msg.ProposedBlock.Data,
		Round: maxRound,
	}

	if !bytes.Equal(maxRoundProposalHash, s.verifier.Keccak(pb.Bytes())) {
		return false
	}

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
