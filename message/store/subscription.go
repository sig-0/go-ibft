package store

import (
	"github.com/rs/xid"
	"github.com/sig-0/go-ibft/message"
)

type subscription[M message.IBFTMessage] struct {
	View         *message.View
	Channel      message.Subscription[M]
	HigherRounds bool
}

func newSubscription[M message.IBFTMessage](view *message.View, higherRounds bool) subscription[M] {
	return subscription[M]{
		View:         view,
		HigherRounds: higherRounds,
		Channel:      make(message.Subscription[M], 1),
	}
}

func (s *subscription[M]) Notify(receiver message.MsgNotificationFn[M]) {
	select {
	case s.Channel <- receiver:
	default: // consumer hasn't used the callback
	}
}

type subscriptions[M message.IBFTMessage] map[string]subscription[M]

func (s *subscriptions[M]) Add(sub subscription[M]) string {
	id := xid.New()
	(*s)[id.String()] = sub

	return id.String()
}

func (s *subscriptions[M]) Remove(id string) {
	close((*s)[id].Channel)
	delete(*s, id)
}

func (s *subscriptions[M]) Notify(notifyFn func(subscription[M])) {
	for _, sub := range *s {
		notifyFn(sub)
	}
}
