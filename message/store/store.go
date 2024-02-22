package store

import (
	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

type message interface {
	types.IBFTMessage

	GetView() *types.View
	GetFrom() []byte
}

// Store is a thread-safe storage for consensus messages with a built-in Feed mechanism
type Store struct {
	ProposalMessages    Collection[*types.MsgProposal]
	PrepareMessages     Collection[*types.MsgPrepare]
	CommitMessages      Collection[*types.MsgCommit]
	RoundChangeMessages Collection[*types.MsgRoundChange]
}

// New returns a new Store instance. NotificationFn added to this store
// have their signatures verified before being included
func New() *Store {
	s := &Store{
		ProposalMessages:    NewCollection[*types.MsgProposal](),
		PrepareMessages:     NewCollection[*types.MsgPrepare](),
		CommitMessages:      NewCollection[*types.MsgCommit](),
		RoundChangeMessages: NewCollection[*types.MsgRoundChange](),
	}

	return s
}

func (s *Store) Feed() ibft.MessageFeed {
	return Feed{s}
}

type Feed struct {
	*Store
}

func (f Feed) ProposalMessages(view *types.View, futureRounds bool) (ibft.Subscription[*types.MsgProposal], func()) {
	return f.Store.ProposalMessages.Subscribe(view, futureRounds)
}

func (f Feed) PrepareMessages(view *types.View, futureRounds bool) (ibft.Subscription[*types.MsgPrepare], func()) {
	return f.Store.PrepareMessages.Subscribe(view, futureRounds)
}

func (f Feed) CommitMessages(view *types.View, futureRounds bool) (ibft.Subscription[*types.MsgCommit], func()) {
	return f.Store.CommitMessages.Subscribe(view, futureRounds)
}

func (f Feed) RoundChangeMessages(view *types.View, futureRounds bool) (ibft.Subscription[*types.MsgRoundChange], func()) {
	return f.Store.RoundChangeMessages.Subscribe(view, futureRounds)
}
