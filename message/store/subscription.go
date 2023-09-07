package store

import (
	"github.com/rs/xid"

	"github.com/madz-lab/go-ibft/message/types"
)

func newSubscription[M msg](view *types.View, higherRounds bool) subscription[M] {
	return subscription[M]{
		View:         view,
		HigherRounds: higherRounds,
		Channel:      make(chan func() []*M, 1),
	}
}

type subscription[M msg] struct {
	View         *types.View
	HigherRounds bool
	Channel      chan func() []*M
}

func (s *subscription[M]) notify(unwrapFn func() []*M) {
	select {
	case s.Channel <- unwrapFn:
	default:
	}
}

func newSubscriptions[M msg]() subscriptions[M] {
	return map[string]subscription[M]{}
}

type subscriptions[M msg] map[string]subscription[M]

func (s *subscriptions[M]) add(sub subscription[M]) string {
	id := xid.New()
	(*s)[id.String()] = sub

	return id.String()
}

func (s *subscriptions[M]) remove(id string) {
	close((*s)[id].Channel)
	delete(*s, id)
}

func (s *subscriptions[M]) notify(notifyFn func(subscription[M])) {
	for _, sub := range *s {
		notifyFn(sub)
	}
}
