package sequencer

import (
	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

type msgCache[M types.IBFTMessage] struct {
	filterFn func(*M) bool
	seen     map[string]struct{}
	filtered []*M
}

func newMsgCache[M types.IBFTMessage](filterFn func(*M) bool) msgCache[M] {
	return msgCache[M]{
		filterFn: filterFn,
		filtered: make([]*M, 0),
		seen:     make(map[string]struct{}),
	}
}

func (c msgCache[M]) Add(messages []*M) msgCache[M] {
	for _, msg := range messages {
		from := any(msg).(ibft.Message).GetFrom() //nolint:forcetypeassert // msg constraint
		if _, ok := c.seen[string(from)]; ok {
			continue
		}

		c.seen[string(from)] = struct{}{}

		if !c.filterFn(msg) {
			continue
		}

		c.filtered = append(c.filtered, msg)
	}

	return c
}

func (c msgCache[M]) Messages() []*M {
	return c.filtered
}
