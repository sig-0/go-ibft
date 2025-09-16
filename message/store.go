package message

import "fmt"

// Store is a thread-safe storage for consensus messages with a built-in sequencer.Feed mechanism
type Store struct {
	ProposalMessages    *Collection[*Proposal]
	PrepareMessages     *Collection[*Prepare]
	CommitMessages      *Collection[*Commit]
	RoundChangeMessages *Collection[*RoundChange]
}

// NewMsgStore returns a new Store instance
func NewMsgStore(messages ...message) *Store {
	s := &Store{
		ProposalMessages:    NewMsgCollection[*Proposal](),
		PrepareMessages:     NewMsgCollection[*Prepare](),
		CommitMessages:      NewMsgCollection[*Commit](),
		RoundChangeMessages: NewMsgCollection[*RoundChange](),
	}

	for _, m := range messages {
		s.Add(m)
	}
	return s
}

// Add includes the message in the store
func (s *Store) Add(msg message) {
	switch msg := msg.(type) {
	case *Proposal:
		s.ProposalMessages.Add(msg)
	case *Prepare:
		s.PrepareMessages.Add(msg)
	case *Commit:
		s.CommitMessages.Add(msg)
	case *RoundChange:
		s.RoundChangeMessages.Add(msg)
	default:
		panic(fmt.Sprintf("unknown message type: %T", msg))
	}
}

// Clear removes all messages from store
func (s *Store) Clear() {
	s.ProposalMessages.Clear()
	s.PrepareMessages.Clear()
	s.CommitMessages.Clear()
	s.RoundChangeMessages.Clear()
}
