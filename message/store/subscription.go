package store

import (
	"github.com/rs/xid"

	"github.com/madz-lab/go-ibft/message/types"
)

type subscription[M types.IBFTMessage] struct {
	View         *types.View
	Channel      types.Subscription[M]
	HigherRounds bool
}

func newSubscription[M types.IBFTMessage](view *types.View, higherRounds bool) subscription[M] {
	return subscription[M]{
		View:         view,
		HigherRounds: higherRounds,
		Channel:      make(types.Subscription[M], 1),
	}
}

func (s *subscription[M]) Notify(receiver types.MsgNotificationFn[M]) {
	select {
	case s.Channel <- receiver:
	default: // consumer hasn't used the callback
	}
}

type subscriptions[M types.IBFTMessage] map[string]subscription[M]

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
