package sequencer

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/madz-lab/go-ibft/message/types"
)

type Verifier interface {
	IsProposer(id []byte, sequence uint64, round uint64) bool
	IsValidator(id []byte, height uint64) bool
	IsValidBlock(block []byte) bool
}

type Validator interface {
	ID() []byte
	Sign([]byte) []byte
	BuildBlock() []byte
}

type Sequencer struct {
	validator Validator

	verifier Verifier

	transport Transport

	quorum Quorum

	cdc Codec

	state state

	round0Duration time.Duration
	wg             sync.WaitGroup

	id []byte
}

func New(val Validator, vrf Verifier, opts ...Option) *Sequencer {
	s := &Sequencer{
		id:             val.ID(),
		validator:      val,
		verifier:       vrf,
		transport:      DummyTransport,
		quorum:         TrueQuorum,
		round0Duration: DefaultRound0Duration,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Sequencer) FinalizeSequence(ctx context.Context, sequence uint64, feed MessageFeed) *types.FinalizedBlock {
	s.state = state{
		currentView: &types.View{
			Sequence: sequence,
			Round:    0,
		},
	}

	c := make(chan *types.FinalizedBlock, 1)
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

func (s *Sequencer) finalize(ctx context.Context, feed MessageFeed) *types.FinalizedBlock {
	for {
		ctxRound, cancelRound := context.WithCancel(ctx)
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
			s.transport.Multicast(s.buildMsgRoundChange())

			println("round timer expired")
		case rcc, ok := <-s.watchForFutureRCC(ctxRound, feed):
			teardown()
			if !ok {
				return nil
			}

			s.state.MoveToRound(rcc.Messages[0].View.Round)
			s.state.roundChangeCertificate = rcc

			println("future rcc")
		case proposal, ok := <-s.watchForFutureProposal(ctxRound, feed):
			teardown()
			if !ok {
				return nil
			}

			s.state.MoveToRound(proposal.View.Round)
			s.state.acceptedProposal = proposal
			s.transport.Multicast(s.buildMsgPrepare())

			println("future proposal")
		case fb, ok := <-s.finalizeBlockInCurrentRound(ctxRound, feed):
			teardown()
			if !ok {
				return nil
			}

			println("finalized block")
			return fb
		}

	}
}

func (s *Sequencer) startRoundTimer(ctx context.Context) <-chan struct{} {
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

func (s *Sequencer) watchForFutureProposal(ctx context.Context, feed MessageFeed) <-chan *types.MsgProposal {
	c := make(chan *types.MsgProposal, 1)

	s.wg.Add(1)
	go func() {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		proposal, err := s.awaitFutureProposal(ctx, feed)
		if err != nil {
			return
		}

		c <- proposal
	}()

	return c
}

func (s *Sequencer) watchForFutureRCC(ctx context.Context, feed MessageFeed) <-chan *types.RoundChangeCertificate {
	c := make(chan *types.RoundChangeCertificate, 1)

	s.wg.Add(1)
	go func() {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		rcc, err := s.awaitFutureRCC(ctx, feed)
		if err != nil {
			return
		}

		c <- rcc
	}()

	return c
}

func (s *Sequencer) finalizeBlockInCurrentRound(ctx context.Context, feed MessageFeed) <-chan *types.FinalizedBlock {
	c := make(chan *types.FinalizedBlock, 1)

	s.wg.Add(1)
	go func() {
		defer func() {
			close(c)
			s.wg.Done()
		}()

		if s.shouldPropose() {
			if err := s.propose(ctx, feed); err != nil {
				return
			}
		}

		if err := s.awaitProposal(ctx, feed); err != nil {
			return
		}

		if err := s.awaitPrepare(ctx, feed); err != nil {
			return
		}

		if err := s.awaitCommit(ctx, feed); err != nil {
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
	return s.verifier.IsProposer(s.id, s.state.CurrentSequence(), s.state.CurrentRound())
}

func (s *Sequencer) propose(ctx context.Context, feed MessageFeed) error {
	block, err := s.buildBlock(ctx, feed)
	if err != nil {
		return err
	}

	msg := s.buildMsgProposal(block)

	s.state.acceptedProposal = msg
	s.transport.Multicast(msg)

	return nil
}

func (s *Sequencer) buildBlock(ctx context.Context, feed MessageFeed) ([]byte, error) {
	if s.state.CurrentRound() == 0 {
		return s.validator.BuildBlock(), nil
	}

	rcc := s.state.roundChangeCertificate
	if rcc == nil {
		// round hop triggered by round timer, justify proposal message with rcc
		rCc, err := s.awaitRCC(ctx, feed)
		if err != nil {
			return nil, err
		}

		rcc = rCc
	}

	block, _ := rcc.HighestRoundBlock()
	if block == nil {
		return s.validator.BuildBlock(), nil
	}

	return block, nil
}

func (s *Sequencer) hash(pb *types.ProposedBlock) []byte {
	return s.cdc.Keccak(pb.Bytes())
}
