package store

import (
	"sync"

	"github.com/rs/xid"

	"github.com/madz-lab/go-ibft/message/types"
)

type subscriptions[M msg] struct {
	mux  sync.RWMutex
	subs map[string]subscription[M]
}

func newSubscriptions[M msg]() subscriptions[M] {
	return subscriptions[M]{
		subs: map[string]subscription[M]{},
	}
}

func (s *subscriptions[M]) add(sub subscription[M]) string {
	s.mux.Lock()
	defer s.mux.Unlock()

	id := xid.New()
	s.subs[id.String()] = sub

	return id.String()
}

func (s *subscriptions[M]) remove(id string) {
	s.mux.Lock()
	defer s.mux.Unlock()

	close(s.subs[id].Channel)
	delete(s.subs, id)
}

func (s *subscriptions[M]) notify(notifyFn func(subscription[M])) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	for _, sub := range s.subs {
		notifyFn(sub)
	}
}

type subscription[M msg] struct {
	View         *types.View
	FutureRounds bool
	Channel      chan func() []*M
}

func newSubscription[M msg](view *types.View, futureRounds bool) subscription[M] {
	return subscription[M]{
		View:         view,
		FutureRounds: futureRounds,
		Channel:      make(chan func() []*M, 1),
	}
}

func (s *subscription[M]) notify(unwrapFn func() []*M) {
	select {
	case s.Channel <- unwrapFn:
	default:
	}
}
