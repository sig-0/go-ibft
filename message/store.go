package message

import (
	"sync"

	"github.com/rs/xid"
)

// Store is a thread-safe storage for consensus messages with a built-in sequencer.Feed mechanism
type Store struct {
	ProposalMessages    Collection[*Proposal]
	PrepareMessages     Collection[*Prepare]
	CommitMessages      Collection[*Commit]
	RoundChangeMessages Collection[*RoundChange]
}

// NewStore returns a new Store instance
func NewStore() *Store {
	return &Store{
		ProposalMessages: Collection[*Proposal]{
			messages: make(map[uint64]set[*Proposal]),
			subs:     make(map[string]subscription[*Proposal]),
		},
		PrepareMessages: Collection[*Prepare]{
			messages: make(map[uint64]set[*Prepare]),
			subs:     make(map[string]subscription[*Prepare]),
		},
		CommitMessages: Collection[*Commit]{
			messages: make(map[uint64]set[*Commit]),
			subs:     make(map[string]subscription[*Commit]),
		},
		RoundChangeMessages: Collection[*RoundChange]{
			messages: make(map[uint64]set[*RoundChange]),
			subs:     make(map[string]subscription[*RoundChange]),
		},
	}
}

type Collection[M message] struct {
	messages    map[uint64]set[M]
	messagesMux sync.RWMutex

	subs    map[string]subscription[M]
	subsMux sync.RWMutex
}

func (c *Collection[M]) Add(m M) {
	// add to store
	func() {
		c.messagesMux.Lock()
		defer c.messagesMux.Unlock()

		s, ok := c.messages[m.GetSequence()]
		if !ok {
			c.messages[m.GetSequence()] = make(set[M])
			s = c.messages[m.GetSequence()]
		}

		s[string(m.GetSignature())] = m
	}()

	// notify subscriptions
	func() {
		c.subsMux.RLock()
		defer c.subsMux.RUnlock()

		for _, sub := range c.subs {
			if sub.sequence != m.GetSequence() {
				continue
			}

			sub.notify(func() []M {
				return c.GetSequence(sub.sequence)
			})
		}
	}()
}

func (c *Collection[M]) Remove(m M) {
	c.messagesMux.Lock()
	defer c.messagesMux.Unlock()

	s, ok := c.messages[m.GetSequence()]
	if !ok {
		return // no message
	}

	delete(s, string(m.GetSignature()))
}

func (c *Collection[M]) GetAll() map[uint64][]M {
	c.messagesMux.RLock()
	defer c.messagesMux.RUnlock()

	res := make(map[uint64][]M, len(c.messages))
	for sequence, set := range c.messages {
		res[sequence] = set.get()
	}

	return res
}

func (c *Collection[M]) GetSequence(sequence uint64) []M {
	c.messagesMux.RLock()
	defer c.messagesMux.RUnlock()

	s, ok := c.messages[sequence]
	if !ok {
		return nil
	}

	return s.get()
}

func (c *Collection[M]) ClearAll() {
	c.messagesMux.Lock()
	defer c.messagesMux.Unlock()

	clear(c.messages)
}

func (c *Collection[M]) ClearSequence(sequence uint64) {
	c.messagesMux.Lock()
	defer c.messagesMux.Unlock()

	clear(c.messages[sequence])
}

func (c *Collection[M]) Subscribe(sequence uint64) (<-chan func() []M, func()) {
	sub := newSubscription[M](sequence)
	id := xid.New().String()

	c.subsMux.Lock()
	c.subs[id] = sub
	c.subsMux.Unlock()

	unsubscribe := func() {
		c.subsMux.Lock()
		defer c.subsMux.Unlock()
		delete(c.subs, id)
	}

	sub.notify(func() []M {
		return c.GetSequence(sequence)
	})

	return sub.ch, unsubscribe
}

type set[M message] map[string]M

func (s set[M]) get() []M {
	res := make([]M, 0, len(s))
	for _, m := range s {
		res = append(res, m)
	}

	return res
}

type subscription[M message] struct {
	sequence uint64
	ch       chan func() []M
}

func newSubscription[M message](sequence uint64) subscription[M] {
	return subscription[M]{
		sequence: sequence,
		ch:       make(chan func() []M, 1),
	}
}

func (s *subscription[M]) notify(receiver func() []M) {
	select {
	case s.ch <- receiver:
	default: // subscriber hasn't consumed the callback
	}
}
