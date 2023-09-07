package sequencer

import (
	ibft "github.com/madz-lab/go-ibft"
	"time"

	"github.com/madz-lab/go-ibft/message/types"
)

var (
	DefaultRound0Duration = time.Second * 10
	DummyTransport        = TransportFn(func(_ types.Msg) {})
	TrueQuorum            = QuorumFn(func(_ []types.Msg) bool { return true })
)

type Option func(*Sequencer)

func WithKeccak(k ibft.Keccak) Option {
	return func(s *Sequencer) {
		s.keccak = k
	}
}

func WithSigRecover(r ibft.SigRecover) Option {
	return func(s *Sequencer) {
		s.recover = r
	}
}

func WithTransport(t Transport) Option {
	return func(s *Sequencer) {
		s.transport = t
	}
}

func WithQuorum(q Quorum) Option {
	return func(s *Sequencer) {
		s.quorum = q
	}
}

func WithRound0Duration(d time.Duration) Option {
	return func(s *Sequencer) {
		s.round0Duration = d
	}
}
