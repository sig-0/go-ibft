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

func New(val ibft.Validator, vrf ibft.Verifier, round0Duration time.Duration) *Sequencer {
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

			s.state.MoveToRound(s.state.CurrentRound() + 1)

			msg := &types.MsgRoundChange{
				View:                        s.state.currentView,
				From:                        s.ID(),
				LatestPreparedProposedBlock: s.state.latestPreparedProposedBlock,
				LatestPreparedCertificate:   s.state.latestPreparedCertificate,
			}

			msg.Signature = s.Sign(msg.Payload())

			ctx.Transport().Multicast(msg)

			println("round timer expired")
		case rcc, ok := <-s.watchForFutureRCC(ctxRound):
			teardown()

			if !ok {
				return nil
			}

			s.state.MoveToRound(rcc.Messages[0].View.Round)
			s.state.roundChangeCertificate = rcc

			println("future rcc")
		case proposal, ok := <-s.watchForFutureProposal(ctxRound):
			teardown()

			if !ok {
				return nil
			}

			s.state.MoveToRound(proposal.View.Round)
			s.state.acceptedProposal = proposal

			msg := &types.MsgPrepare{
				From:      s.ID(),
				View:      s.state.currentView,
				BlockHash: s.state.AcceptedBlockHash(),
			}

			msg.Signature = s.Sign(msg.Payload())

			ctx.Transport().Multicast(msg)

		case fb, ok := <-s.finalizeBlockInCurrentRound(ctxRound):
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

func (s *Sequencer) watchForFutureProposal(ctx ibft.Context) <-chan *types.MsgProposal {
	s.wg.Add(1)

	c := make(chan *types.MsgProposal, 1)
	go func() {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		proposal, err := s.awaitFutureProposal(ctx)
		if err != nil {
			return
		}

		c <- proposal
	}()

	return c
}

func (s *Sequencer) watchForFutureRCC(ctx ibft.Context) <-chan *types.RoundChangeCertificate {
	s.wg.Add(1)

	c := make(chan *types.RoundChangeCertificate, 1)
	go func() {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		rcc, err := s.awaitFutureRCC(ctx)
		if err != nil {
			return
		}

		c <- rcc
	}()

	return c
}

func (s *Sequencer) finalizeBlockInCurrentRound(ctx ibft.Context) <-chan *types.FinalizedBlock {
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

		if err := s.awaitProposal(ctx); err != nil {
			return
		}

		if err := s.awaitPrepare(ctx); err != nil {
			return
		}

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

	pb := &types.ProposedBlock{
		Block: block,
		Round: s.state.CurrentRound(),
	}

	msg := &types.MsgProposal{
		From:                   s.ID(),
		View:                   s.state.currentView,
		ProposedBlock:          pb,
		BlockHash:              ctx.Keccak().Hash(pb.Bytes()),
		RoundChangeCertificate: s.state.roundChangeCertificate,
	}

	msg.Signature = s.Sign(msg.Payload())

	s.state.acceptedProposal = msg

	ctx.Transport().Multicast(msg)

	return nil
}

func (s *Sequencer) buildBlock(ctx ibft.Context) ([]byte, error) {
	if s.state.CurrentRound() == 0 {
		return s.BuildBlock(s.state.CurrentSequence()), nil
	}

	rcc := s.state.roundChangeCertificate
	if rcc == nil {
		// round jump triggered by round timer, justify proposal with rcc
		rCc, err := s.awaitRCC(ctx)
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
