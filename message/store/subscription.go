package store

import (
	"github.com/rs/xid"
	"github.com/sig-0/go-ibft/message"
)

type subscription[M message.Message] struct {
	sub          chan func() []M
	sequence     uint64
	round        uint64
	higherRounds bool
}

func newSubscription[M message.Message](sequence, round uint64, higherRounds bool) subscription[M] {
	return subscription[M]{
		sequence:     sequence,
		round:        round,
		higherRounds: higherRounds,
		sub:          make(chan func() []M, 1),
	}
}

func (s *subscription[M]) notify(receiver func() []M) {
	select {
	case s.sub <- receiver:
	default: // subscriber hasn't consumed the callback
	}
}

type subscriptions[M message.Message] map[string]subscription[M]

func (s *subscriptions[M]) add(sub subscription[M]) string {
	id := xid.New()
	(*s)[id.String()] = sub

	return id.String()
}

func (s *subscriptions[M]) remove(id string) {
	close((*s)[id].sub)
	delete(*s, id)
}

func (s *subscriptions[M]) Notify(notifyFn func(subscription[M])) {
	for _, sub := range *s {
		notifyFn(sub)
	}
}
