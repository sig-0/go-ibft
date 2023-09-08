package store

import (
	"github.com/madz-lab/go-ibft/message/types"
)

// Feed is the types.MsgFeed implementation object
type Feed struct {
	*Store
}

// Proposal returns the proposal subscription channel
func (f Feed) Proposal(view *types.View, futureRounds bool) (<-chan func() []*types.MsgProposal, func()) {
	return f.Store.proposal.Subscribe(view, futureRounds)
}

// Prepare returns the prepare subscription channel
func (f Feed) Prepare(view *types.View, futureRounds bool) (<-chan func() []*types.MsgPrepare, func()) {
	return f.Store.prepare.Subscribe(view, futureRounds)
}

// Commit returns the commit subscription channel
func (f Feed) Commit(view *types.View, futureRounds bool) (<-chan func() []*types.MsgCommit, func()) {
	return f.Store.commit.Subscribe(view, futureRounds)
}

// RoundChange returns the round change subscription channel
func (f Feed) RoundChange(view *types.View, futureRounds bool) (<-chan func() []*types.MsgRoundChange, func()) {
	return f.Store.roundChange.Subscribe(view, futureRounds)
}
