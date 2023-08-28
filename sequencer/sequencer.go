package sequencer

import (
	"bytes"
	"context"
	"github.com/madz-lab/go-ibft/message/types"
	"math"
	"sync"
	"time"
)

type MessageFeed interface {
	SubscribeToProposalMessages(view *types.View, higherRounds bool) (<-chan func() []*types.MsgProposal, func())
	SubscribeToPrepareMessages(view *types.View, higherRounds bool) (<-chan func() []*types.MsgPrepare, func())
	SubscribeToCommitMessages(view *types.View, higherRounds bool) (<-chan func() []*types.MsgCommit, func())
	SubscribeToRoundChangeMessages(view *types.View, higherRounds bool) (<-chan func() []*types.MsgRoundChange, func())
}

type Verifier interface {
	Keccak([]byte) []byte
	IsValidBlock([]byte) bool
	IsProposer(view *types.View, id []byte) bool
	RecoverFrom(data []byte, sig []byte) []byte
}

type Validator interface {
	ID() []byte
	Sign([]byte) []byte
	BuildBlock() []byte
}

type FinalizedBlock struct {
	Block       []byte
	Round       uint64
	CommitSeals [][]byte
}

type Sequencer struct {
	validator Validator

	verifier Verifier

	transport Transport

	quorum Quorum

	state state

	round0Duration time.Duration
	wg             sync.WaitGroup

	id []byte
}

func New(
	validator Validator,
	verifier Verifier,
	round0Duration time.Duration,
) *Sequencer {
	return &Sequencer{
		id:             validator.ID(),
		validator:      validator,
		verifier:       verifier,
		round0Duration: round0Duration,
	}
}

type state struct {
	currentView                 *types.View
	acceptedProposal            *types.MsgProposal
	latestPreparedProposedBlock *types.ProposedBlock
	latestPreparedCertificate   *types.PreparedCertificate
	finalizedBlock              *FinalizedBlock
}

func (s *Sequencer) WithTransport(t Transport) {
	s.transport = t
}

func (s *Sequencer) WithQuorum(q Quorum) {
	s.quorum = q
}

func (s *Sequencer) FinalizeSequence(ctx context.Context, sequence uint64, feed MessageFeed) *FinalizedBlock {
	s.state = state{
		currentView: &types.View{Sequence: sequence, Round: 0},
	}

	c := make(chan *FinalizedBlock)
	go func() {
		defer close(c)

		fb := s.finalize(ctx, feed)
		if fb == nil {
			return
		}

		c <- fb
	}()

	select {
	case <-ctx.Done():
		<-c // wait for finalize to return
		return nil
	case fb := <-c:
		return fb
	}
}

func (s *Sequencer) finalize(ctx context.Context, feed MessageFeed) *FinalizedBlock {
	for {
		view := s.state.currentView

		ctxRound, cancelRound := context.WithCancel(ctx)
		teardown := func() {
			cancelRound()
			s.wg.Wait()
		}

		select {
		case _, ok := <-s.startRoundTimer(ctxRound, view):
			teardown()
			if !ok {
				return nil
			}

		//	todo
		case _, ok := <-s.watchForFutureRCC(ctxRound, view):
			teardown()
			if !ok {
				return nil
			}

		//	todo
		case _, ok := <-s.watchForFutureProposal(ctxRound, view):
			teardown()
			if !ok {
				return nil
			}

		//	todo
		case _, ok := <-s.runRound(ctxRound, view, feed):
			teardown()
			if !ok {
				return nil
			}

			return s.state.finalizedBlock
		}
	}
}

func (s *Sequencer) startRoundTimer(ctx context.Context, view *types.View) <-chan struct{} {
	s.wg.Add(1)
	c := make(chan struct{})

	go func() {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		roundTimer := time.NewTimer(s.getRoundDuration(view.Round))

		select {
		case <-ctx.Done():
			roundTimer.Stop()
		case <-roundTimer.C:
			c <- struct{}{}
		}
	}()

	return c
}

func (s *Sequencer) getRoundDuration(round uint64) time.Duration {
	return s.round0Duration * time.Duration(math.Pow(2, float64(round)))
}

func (s *Sequencer) watchForFutureProposal(ctx context.Context, view *types.View) <-chan *types.MsgProposal {
	return nil
}

func (s *Sequencer) watchForFutureRCC(ctx context.Context, view *types.View) <-chan *types.RoundChangeCertificate {
	return nil
}

func (s *Sequencer) runRound(ctx context.Context, view *types.View, feed MessageFeed) <-chan struct{} {
	s.wg.Add(1)
	c := make(chan struct{})

	go func() {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		// if proposer
		if s.verifier.IsProposer(view, s.id) {
			if err := s.propose(ctx, view); err != nil {
				return
			}
		}

		// new round
		if err := s.waitForProposal(ctx, view, feed); err != nil {
			return
		}

		// prepare
		if err := s.waitForPrepare(ctx, view, feed); err != nil {
			return
		}

		// commit
		if err := s.waitForCommit(ctx, view, feed); err != nil {
			return
		}

		c <- struct{}{}
	}()

	return c
}

func (s *Sequencer) propose(ctx context.Context, view *types.View) error {
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

	if !bytes.Equal(msg.GetFrom(), s.verifier.RecoverFrom(s.state.acceptedProposal.GetProposalHash(), msg.GetCommitSeal())) {
		return false
	}

	return true
}
