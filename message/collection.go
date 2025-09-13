package message

import (
	"sync"
)

type Collection[M message] struct {
	msgCollection[M]
	subscriptions[M]

	collectionMux, subscriptionMux sync.RWMutex
}

func NewMsgCollection[M message]() *Collection[M] {
	return &Collection[M]{
		msgCollection: msgCollection[M]{},
		subscriptions: subscriptions[M]{},
	}
}

func (c *Collection[M]) Clear() {
	c.collectionMux.Lock()
	defer c.collectionMux.Unlock()

	clear(c.msgCollection)
}

func (c *Collection[M]) Subscribe(sequence, round uint64, higherRounds bool) (chan func() []M, func()) {
	sub := newSubscription[M](sequence, round, higherRounds)
	unregister := c.registerSubscription(sub)

	sub.notify(c.getNotificationFn(sequence, round, higherRounds))

	return sub.sub, unregister
}

func (c *Collection[M]) registerSubscription(sub subscription[M]) func() {
	c.subscriptionMux.Lock()
	defer c.subscriptionMux.Unlock()

	id := c.subscriptions.add(sub)

	return func() {
		c.subscriptionMux.Lock()
		defer c.subscriptionMux.Unlock()

		c.subscriptions.remove(id)
	}
}

func (c *Collection[M]) Add(msg M) {
	c.collectionMux.Lock()
	defer c.collectionMux.Unlock()

	c.msgCollection.add(msg)

	c.subscriptionMux.RLock()
	defer c.subscriptionMux.RUnlock()

	seq, round := msg.GetSequence(), msg.GetRound()

	c.subscriptions.Notify(func(sub subscription[M]) {
		// match the sequence
		if seq != sub.sequence {
			return
		}

		// exclude lower rounds
		if round < sub.round {
			return
		}

		sub.notify(c.getNotificationFn(sub.sequence, sub.round, sub.higherRounds))
	})
}

func (c *Collection[M]) Get(sequence, round uint64) []M {
	c.collectionMux.RLock()
	defer c.collectionMux.RUnlock()

	return c.msgCollection.loadSet(sequence, round).Messages()
}

func (c *Collection[M]) getNotificationFn(sequence, round uint64, higherRounds bool) func() []M {
	return func() []M {
		c.collectionMux.RLock()
		defer c.collectionMux.RUnlock()

		if !higherRounds {
			return c.msgCollection.get(sequence, round)
		}

		return c.msgCollection.getMessagesWithHighestRoundNumber(sequence, round)
	}
}

type msgCollection[M message] map[uint64]map[uint64]msgSet[M]

func (c *msgCollection[M]) add(msg M) {
	var (
		sequence = msg.GetSequence()
		round    = msg.GetRound()
		sender   = msg.GetSender()
	)

	c.loadOrStoreSet(sequence, round)[string(sender)] = msg
}

func (c *msgCollection[M]) loadOrStoreSet(sequence, round uint64) msgSet[M] {
	sameSequenceMessages, ok := (*c)[sequence]
	if !ok {
		(*c)[sequence] = map[uint64]msgSet[M]{}
		sameSequenceMessages = (*c)[sequence]
	}

	set, ok := sameSequenceMessages[round]
	if !ok {
		(*c)[sequence][round] = msgSet[M]{}
		set = (*c)[sequence][round]
	}

	return set
}

func (c *msgCollection[M]) get(sequence, round uint64) []M {
	return c.loadSet(sequence, round).Messages()
}

func (c *msgCollection[M]) loadSet(sequence, round uint64) msgSet[M] {
	sameSequenceMessages, ok := (*c)[sequence]
	if !ok {
		return nil
	}

	set, ok := sameSequenceMessages[round]
	if !ok {
		return nil
	}

	return set
}

func (c *msgCollection[M]) getMessagesWithHighestRoundNumber(sequence, round uint64) []M {
	maxRound := round
	for round := range (*c)[sequence] {
		if maxRound >= round {
			continue
		}

		maxRound = round
	}

	return c.get(sequence, maxRound)
}

type msgSet[M message] map[string]M

func (s msgSet[M]) Messages() []M {
	messages := make([]M, 0, len(s))
	for _, msg := range s {
		messages = append(messages, msg)
	}

	return messages
}
