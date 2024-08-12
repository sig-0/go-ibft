package store

import (
	"github.com/rs/xid"
	"github.com/sig-0/go-ibft/message"
)

type subscription[M message.IBFTMessage] struct {
	sub          message.Subscription[M]
	sequence     uint64
	round        uint64
	higherRounds bool
}

func newSubscription[M message.IBFTMessage](sequence, round uint64, higherRounds bool) subscription[M] {
	return subscription[M]{
		sequence:     sequence,
		round:        round,
		higherRounds: higherRounds,
		sub:          make(message.Subscription[M], 1),
	}
}

func (s *subscription[M]) notify(receiver message.MsgNotificationFn[M]) {
	select {
	case s.sub <- receiver:
	default: // subscriber hasn't consumed the callback
	}
}

type subscriptions[M message.IBFTMessage] map[string]subscription[M]

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
