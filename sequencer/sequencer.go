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
	Keccak(data []byte) []byte
	RecoverFrom(data []byte, sig []byte) []byte // todo: remove with IsValidator and move to message store
	IsProposer(view *types.View, id []byte) bool
	IsValidBlock(block []byte) bool
	IsValidator(id []byte, height uint64) bool
}

type Validator interface {
	ID() []byte
	Sign(data []byte) []byte
	BuildBlock() []byte
}

type FinalizedBlock struct {
	Block       []byte
	CommitSeals [][]byte
	Round       uint64
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
	roundChangeCertificate      *types.RoundChangeCertificate
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

			s.state.currentView.Round++
			s.state.acceptedProposal = nil

			s.multicastRoundChangeMessage(s.state.currentView)
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

		if s.verifier.IsProposer(view, s.id) {
			if err := s.propose(ctx, view, feed); err != nil {
				return
			}
		}

		// proposal.go
		if err := s.waitForProposal(ctx, view, feed); err != nil {
			return
		}

		// prepare.go
		if err := s.waitForPrepare(ctx, view, feed); err != nil {
			return
		}

		// commit.go
		if err := s.waitForCommit(ctx, view, feed); err != nil {
			return
		}

		c <- struct{}{}
	}()

	return c
}

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
