package sequencer

import (
	"math"
	"sync"
	"time"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

type Sequencer struct {
	ibft.Validator
	ibft.Verifier

	wg             sync.WaitGroup
	state          state
	round0Duration time.Duration
}

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

func (s *Sequencer) startRoundTimer(ctx ibft.Context) <-chan struct{} {
	c := make(chan struct{}, 1)

	s.wg.Add(1)

	go func() {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		roundTimer := s.getRoundTimer()

		select {
		case <-ctx.Done():
			roundTimer.Stop()
		case <-roundTimer.C:
			c <- struct{}{}
		}
	}()

	return c
}

func (s *Sequencer) awaitHigherRoundProposal(ctx ibft.Context) <-chan *types.MsgProposal {
	s.wg.Add(1)

	c := make(chan *types.MsgProposal, 1)
	go func() {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		proposal, err := s.awaitProposal(ctx, true)
		if err != nil {
			return
		}

		c <- proposal
	}()

	return c
}

func (s *Sequencer) awaitHigherRoundRCC(ctx ibft.Context) <-chan *types.RoundChangeCertificate {
	s.wg.Add(1)

	c := make(chan *types.RoundChangeCertificate, 1)
	go func() {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		rcc, err := s.awaitRCC(ctx, true)
		if err != nil {
			return
		}

		c <- rcc
	}()

	return c
}

func (s *Sequencer) awaitFinalizedBlockInCurrentRound(ctx ibft.Context) <-chan *types.FinalizedBlock {
	s.wg.Add(1)

	c := make(chan *types.FinalizedBlock, 1)
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

func (s *Sequencer) getRoundTimer() *time.Timer {
	return time.NewTimer(s.round0Duration * time.Duration(math.Pow(2, float64(s.state.CurrentRound()))))
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

	rcc := s.state.roundChangeCertificate
	if rcc == nil {
		// round jump triggered by round timer, justify proposal with rcc
		rCc, err := s.awaitRCC(ctx, false)
		if err != nil {
			return nil, err
		}

		rcc = rCc
	}

	block, _ := rcc.HighestRoundBlock()
	if block == nil {
		return s.BuildBlock(s.state.CurrentSequence()), nil
	}

	return block, nil
}
