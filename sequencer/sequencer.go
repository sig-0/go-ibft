package sequencer

import (
	"math"
	"sync"
	"time"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

// Sequencer is an IBFT-context based block finalizer of a consensus actor. Its only purpose
// is to determine if consensus is reached in a particular sequence (height)
// and output the finalized block. Because of its simple API it's envisioned to
// work alongside a block syncing process
type Sequencer struct {
	ibft.Validator
	ibft.Verifier

	state          state
	wg             sync.WaitGroup
	round0Duration time.Duration
}

// New instantiates a new Sequencer object
func New(
	val ibft.Validator,
	vrf ibft.Verifier,
	round0Duration time.Duration,
) *Sequencer {
	s := &Sequencer{
		Validator:      val,
		Verifier:       vrf,
		round0Duration: round0Duration,
	}

	return s
}

// FinalizeSequence runs the block finalization loop. This method returns a non-nil value only if
// consensus is reached for the given sequence. Otherwise, it runs forever until cancelled by the caller
func (s *Sequencer) FinalizeSequence(ctx ibft.Context, sequence uint64) *types.FinalizedBlock {
	s.state = state{currentView: &types.View{
		Sequence: sequence,
		Round:    0,
	}}

	c := make(chan *types.FinalizedBlock, 1)
	go func() {
		defer close(c)

		fb := s.finalize(ctx)
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

// finalize starts the round runner loop. In each round (loop iteration), 4 processes run in parallel.
// This method returns only if the block finalization algorithm is complete or if the caller cancelled the context
func (s *Sequencer) finalize(ctx ibft.Context) *types.FinalizedBlock {
	for {
		ctxRound, cancelRound := ctx.WithCancel()
		teardown := func() {
			cancelRound()
			s.wg.Wait()
		}

		select {
		case _, ok := <-s.startRoundTimer(ctxRound):
			teardown()

			if !ok {
				return nil
			}

			s.state.MoveToNextRound()
			s.multicastRoundChange(ctx)

		case rcc, ok := <-s.awaitHigherRoundRCC(ctxRound):
			teardown()

			if !ok {
				return nil
			}

			s.state.AcceptRCC(rcc)

		case proposal, ok := <-s.awaitHigherRoundProposal(ctxRound):
			teardown()

			if !ok {
				return nil
			}

			s.state.AcceptProposal(proposal)
			s.multicastPrepare(ctx)

		case fb, ok := <-s.awaitFinalizedBlockInCurrentRound(ctxRound):
			teardown()

			if !ok {
				return nil
			}

			return fb
		}
	}
}

// startRoundTimer starts the round timer of the current round
func (s *Sequencer) startRoundTimer(ctx ibft.Context) <-chan struct{} {
	c := make(chan struct{}, 1)

	s.wg.Add(1)

	go func(view *types.View) {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		roundTimer := s.getRoundTimer(view.Round)

		select {
		case <-ctx.Done():
			roundTimer.Stop()
		case <-roundTimer.C:
			c <- struct{}{}
		}

	}(s.state.CurrentView())

	return c
}

// awaitHigherRoundProposal listens for proposal messages from rounds higher than the current
func (s *Sequencer) awaitHigherRoundProposal(ctx ibft.Context) <-chan *types.MsgProposal {
	c := make(chan *types.MsgProposal, 1)

	s.wg.Add(1)

	go func(view *types.View) {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		proposal, err := s.awaitProposal(ctx, view, true)
		if err != nil {
			return
		}

		c <- proposal

	}(s.state.CurrentView())

	return c
}

// awaitHigherRoundRCC listens for round change certificates from rounds higher than the current
func (s *Sequencer) awaitHigherRoundRCC(ctx ibft.Context) <-chan *types.RoundChangeCertificate {
	c := make(chan *types.RoundChangeCertificate, 1)

	s.wg.Add(1)

	go func(view *types.View) {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		rcc, err := s.awaitRCC(ctx, view, true)
		if err != nil {
			return
		}

		c <- rcc

	}(s.state.CurrentView())

	return c
}

// awaitFinalizedBlockInCurrentRound starts the block finalization algorithm for the current round
func (s *Sequencer) awaitFinalizedBlockInCurrentRound(ctx ibft.Context) <-chan *types.FinalizedBlock {
	c := make(chan *types.FinalizedBlock, 1)

	s.wg.Add(1)

	go func() {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		if s.shouldPropose() {
			if err := s.propose(ctx); err != nil {
				return
			}
		}

		if !s.state.ProposalAccepted() {
			if err := s.awaitCurrentRoundProposal(ctx); err != nil {
				return
			}

			s.multicastPrepare(ctx)
		}

		if err := s.awaitPrepare(ctx); err != nil {
			return
		}

		s.multicastCommit(ctx)

		if err := s.awaitCommit(ctx); err != nil {
			return
		}

		c <- s.state.FinalizedBlock()

	}()

	return c
}

func (s *Sequencer) getRoundTimer(round uint64) *time.Timer {
	return time.NewTimer(s.round0Duration * time.Duration(math.Pow(2, float64(round))))
}

func (s *Sequencer) shouldPropose() bool {
	return s.IsProposer(s.ID(), s.state.CurrentSequence(), s.state.CurrentRound())
}

func (s *Sequencer) propose(ctx ibft.Context) error {
	block, err := s.buildBlock(ctx)
	if err != nil {
		return err
	}

	s.multicastProposal(ctx, block)

	return nil
}

func (s *Sequencer) buildBlock(ctx ibft.Context) ([]byte, error) {
	if s.state.CurrentRound() == 0 {
		return s.BuildBlock(s.state.CurrentSequence()), nil
	}

	if s.state.roundChangeCertificate == nil {
		// round jump triggered by round timer, justify proposal with rcc
		rCc, err := s.awaitRCC(ctx, s.state.currentView, false)
		if err != nil {
			return nil, err
		}

		s.state.roundChangeCertificate = rCc
	}

	rcc := s.state.roundChangeCertificate

	block, _ := rcc.HighestRoundBlock()
	if block == nil {
		return s.BuildBlock(s.state.CurrentSequence()), nil
	}

	return block, nil
}
