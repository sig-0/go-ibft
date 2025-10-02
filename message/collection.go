package message

import (
	"sync"
)

type Collection[M message] interface {
	Add(M)
	Remove(M)
	GetAll() map[uint64][]M
	GetAllInSequence(uint64) []M
	ClearAll()
	ClearAllInSequence(uint64)
	Subscribe(uint64) <-chan func() ([]M, func())
}

type set[M message] map[string]M

type coll[M message] struct {
	messages    map[uint64]set[M]
	messagesMux sync.RWMutex

	subs    map[string]subscription[M]
	subsMux sync.RWMutex
}

func (c *coll[M]) Add(m M) {
	c.messagesMux.Lock()
	defer c.messagesMux.Unlock()

	s, ok := c.messages[m.GetSequence()]
	if !ok {
		c.messages[m.GetSequence()] = make(set[M])
		s = c.messages[m.GetSequence()]
	}

	s[string(m.GetSignature())] = m

	c.subsMux.RLock()
	defer c.subsMux.RUnlock()
	for _, sub := range c.subs {
		if sub.sequence != m.GetSequence() {
			continue
		}

		sub.sub <- func() []M {
			return c.GetAllInSequence(sub.sequence)
		}
	}
}

func (c *coll[M]) Remove(m M) {
	c.messagesMux.Lock()
	defer c.messagesMux.Unlock()

	s, ok := c.messages[m.GetSequence()]
	if !ok {
		return // no message here
	}

	delete(s, string(m.GetSignature()))
}

func (c *coll[M]) GetAll() map[uint64][]M {
	c.messagesMux.RLock()
	defer c.messagesMux.RUnlock()

	res := make(map[uint64][]M)
	for sequence, set := range c.messages {
		for _, msg := range set {
			res[sequence] = append(res[sequence], msg)
		}
	}

	return res
}

func (c *coll[M]) GetAllInSequence(sequence uint64) []M {
	c.messagesMux.RLock()
	defer c.messagesMux.RUnlock()

	res := make([]M, 0, len(c.messages[sequence]))
	for _, msg := range c.messages[sequence] {
		res = append(res, msg)
	}

	return res
}

func (c *coll[M]) ClearAll() {
	c.messagesMux.Lock()
	defer c.messagesMux.Unlock()

	clear(c.messages)
}

func (c *coll[M]) ClearAllInSequence(sequence uint64) {
	c.messagesMux.Lock()
	defer c.messagesMux.Unlock()

	clear(c.messages[sequence])
}

func (c *coll[M]) Subscribe(u uint64) <-chan func() ([]M, func()) {
	//TODO implement me
	panic("implement me")
}

type Colllection[M message] struct {
	msgCollection[M]
	subscriptions[M]

	collectionMux, subscriptionMux sync.RWMutex
}

func NewMsgCollection[M message]() *Colllection[M] {
	return &Colllection[M]{
		msgCollection: msgCollection[M]{},
		subscriptions: subscriptions[M]{},
	}
}

func (c *Colllection[M]) Clear() {
	c.collectionMux.Lock()
	defer c.collectionMux.Unlock()

	clear(c.msgCollection)
}

func (c *Colllection[M]) Subscribe(sequence, round uint64, higherRounds bool) (chan func() []M, func()) {
	sub := newSubscription[M](sequence, round, higherRounds)
	unregister := c.registerSubscription(sub)

	sub.notify(c.getNotificationFn(sequence, round, higherRounds))

	return sub.sub, unregister
}

func (c *Colllection[M]) registerSubscription(sub subscription[M]) func() {
	c.subscriptionMux.Lock()
	defer c.subscriptionMux.Unlock()

	id := c.subscriptions.add(sub)

	return func() {
		c.subscriptionMux.Lock()
		defer c.subscriptionMux.Unlock()

		c.subscriptions.remove(id)
	}
}

func (c *Colllection[M]) Add(msg M) {
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

func (c *Colllection[M]) Get(sequence, round uint64) []M {
	c.collectionMux.RLock()
	defer c.collectionMux.RUnlock()

	return c.msgCollection.loadSet(sequence, round).Messages()
}

func (c *Colllection[M]) getNotificationFn(sequence, round uint64, higherRounds bool) func() []M {
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
