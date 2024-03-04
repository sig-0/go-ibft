package store

import (
	"github.com/madz-lab/go-ibft/sequencer"
	"sync"

	"github.com/madz-lab/go-ibft/message/types"
)

type Collection[M message] interface {
	AddMessage(msg M)
	GetMessages(view *types.View) []M
	RemoveMessages(view *types.View)

	Subscribe(view *types.View, higherRounds bool) (sequencer.Subscription[M], func())
}

type syncCollection[M message] struct {
	msgCollection[M]
	subscriptions[M]

	collectionMux, subscriptionMux sync.RWMutex
}

func NewCollection[M message]() Collection[M] {
	return &syncCollection[M]{
		msgCollection: msgCollection[M]{},
		subscriptions: subscriptions[M]{},
	}
}

func (c *syncCollection[M]) Subscribe(view *types.View, higherRounds bool) (sequencer.Subscription[M], func()) {
	sub := newSubscription[M](view, higherRounds)
	unregister := c.registerSubscription(sub)

	sub.Notify(c.unwrapMessagesFn(view, higherRounds))

	return sub.Channel, unregister
}

func (c *syncCollection[M]) registerSubscription(sub subscription[M]) func() {
	c.subscriptionMux.Lock()
	defer c.subscriptionMux.Unlock()

	id := c.subscriptions.Add(sub)

	return func() {
		c.subscriptionMux.Lock()
		defer c.subscriptionMux.Unlock()

		c.subscriptions.Remove(id)
	}
}

func (c *syncCollection[M]) AddMessage(msg M) {
	add := func() {
		c.collectionMux.Lock()
		defer c.collectionMux.Unlock()

		c.msgCollection.add(msg)
	}

	notify := func() {
		c.subscriptionMux.RLock()
		defer c.subscriptionMux.RUnlock()

		view := msg.GetView()

		c.subscriptions.Notify(func(sub subscription[M]) {
			// match the sequence
			if view.Sequence != sub.View.Sequence {
				return
			}

			// exclude lower rounds
			if view.Round < sub.View.Round {
				return
			}

			sub.Notify(c.unwrapMessagesFn(sub.View, sub.HigherRounds))
		})
	}

	add()
	notify()
}

func (c *syncCollection[M]) GetMessages(view *types.View) []M {
	c.collectionMux.RLock()
	defer c.collectionMux.RUnlock()

	return c.msgCollection.loadSet(view).Messages()
}

func (c *syncCollection[M]) unwrapMessagesFn(view *types.View, higherRounds bool) sequencer.NotificationFn[M] {
	return func() []M {
		c.collectionMux.RLock()
		defer c.collectionMux.RUnlock()

		if !higherRounds {
			return c.msgCollection.get(view)
		}

		return c.msgCollection.getMessagesWithHighestRoundNumber(view)
	}
}

func (c *syncCollection[M]) RemoveMessages(view *types.View) {
	c.collectionMux.Lock()
	defer c.collectionMux.Unlock()

	c.msgCollection.remove(view)
}

type msgCollection[M message] map[uint64]map[uint64]msgSet[M]

func (c *msgCollection[M]) add(msg M) {
	c.loadOrStoreSet(msg.GetView())[string(msg.GetSender())] = msg
}

func (c *msgCollection[M]) loadOrStoreSet(view *types.View) msgSet[M] {
	sameSequenceMessages, ok := (*c)[view.Sequence]
	if !ok {
		(*c)[view.Sequence] = map[uint64]msgSet[M]{}
		sameSequenceMessages = (*c)[view.Sequence]
	}

	set, ok := sameSequenceMessages[view.Round]
	if !ok {
		(*c)[view.Sequence][view.Round] = msgSet[M]{}
		set = (*c)[view.Sequence][view.Round]
	}

	return set
}

func (c *msgCollection[M]) get(view *types.View) []M {
	return c.loadSet(view).Messages()
}

func (c *msgCollection[M]) loadSet(view *types.View) msgSet[M] {
	sameSequenceMessages, ok := (*c)[view.Sequence]
	if !ok {
		return nil
	}

	set, ok := sameSequenceMessages[view.Round]
	if !ok {
		return nil
	}

	return set
}

func (c *msgCollection[M]) getMessagesWithHighestRoundNumber(view *types.View) []M {
	maxRound := view.Round
	for round := range (*c)[view.Sequence] {
		if maxRound >= round {
			continue
		}

		maxRound = round
	}

	return c.get(&types.View{Sequence: view.Sequence, Round: maxRound})
}

func (c *msgCollection[M]) remove(view *types.View) {
	_, ok := (*c)[view.Sequence]
	if !ok {
		return
	}

	_, ok = (*c)[view.Sequence][view.Round]
	if !ok {
		return
	}

	delete((*c)[view.Sequence], view.Round)
}

type msgSet[M types.IBFTMessage] map[string]M

func (s msgSet[M]) Messages() []M {
	messages := make([]M, 0, len(s))
	for _, msg := range s {
		messages = append(messages, msg)
	}

	return messages
}
