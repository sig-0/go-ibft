package sequencer

import (
	"github.com/madz-lab/go-ibft/message/types"
)

type msg interface {
	types.MsgProposal | types.MsgPrepare | types.MsgCommit | types.MsgRoundChange
}

type msgCache[M msg] struct {
	filterFn func(*M) bool
	seen     map[string]struct{}
	filtered []*M
}

func newMsgCache[M msg](filterFn func(*M) bool) msgCache[M] {
	return msgCache[M]{
		filterFn: filterFn,
		filtered: make([]*M, 0),
		seen:     make(map[string]struct{}),
	}
}

func (c msgCache[M]) Add(messages []*M) msgCache[M] {
	for _, msg := range messages {
		from := any(msg).(types.Msg).GetFrom()
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
