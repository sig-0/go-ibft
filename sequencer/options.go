package sequencer

import (
	"time"

	"github.com/madz-lab/go-ibft/message/types"
)

var (
	DefaultRound0Duration = time.Second * 10
	DummyTransport        = TransportFn(func(_ types.Msg) {})
	TrueQuorum            = QuorumFn(func(_ []types.Msg) bool { return true })
)

type Option func(*Sequencer)

func WithCodec(cdc Codec) Option {
	return func(s *Sequencer) {
		s.cdc = cdc
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
