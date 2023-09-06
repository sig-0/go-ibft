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

	delete(s.subs, id)
}

func (s *subscriptions[M]) notify(view *types.View, unwrap func() []*M) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	for _, sub := range s.subs {
		if sub.View.Sequence != view.Sequence {
			continue
		}

		if sub.View.Round > view.Round && !sub.FutureRounds {
			continue
		}

		if sub.View.Round != view.Round {
			continue
		}

		select {
		case sub.Channel <- unwrap:
		default:
		}
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

type Feed struct {
	*Store
}

func (f Feed) SubscribeToProposalMessages(view *types.View, futureRounds bool) (<-chan func() []*types.MsgProposal, func()) {
	sub := newSubscription[types.MsgProposal](view, futureRounds)
	id := f.Store.proposalSubs.add(sub)
	sub.Channel <- func() []*types.MsgProposal {
		return f.Store.GetProposalMessages(view)
	}

	return sub.Channel, func() { f.Store.proposalSubs.remove(id) }
}

func (f Feed) SubscribeToPrepareMessages(view *types.View, futureRounds bool) (<-chan func() []*types.MsgPrepare, func()) {
	//TODO implement me
	panic("implement me")
}

func (f Feed) SubscribeToCommitMessages(view *types.View, futureRounds bool) (<-chan func() []*types.MsgCommit, func()) {
	//TODO implement me
	panic("implement me")
}

func (f Feed) SubscribeToRoundChangeMessages(view *types.View, futureRounds bool) (<-chan func() []*types.MsgRoundChange, func()) {
	//TODO implement me
	panic("implement me")
}
