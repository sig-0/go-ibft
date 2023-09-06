package types

type Msg interface {
	Bytes() []byte
	GetFrom() []byte
	Payload() []byte
}

type ibft interface {
	MsgProposal | MsgPrepare | MsgCommit | MsgRoundChange
}

func Filter[M ibft](messages []*M, filter func(*M) bool) []*M {
	filtered := make([]*M, 0, len(messages))
	for _, msg := range messages {
		if filter(msg) {
			filtered = append(filtered, msg)
		}
	}

	return filtered
}

func ToMsg[M ibft](messages []*M) []Msg {
	wrapped := make([]Msg, 0, len(messages))
	for _, msg := range messages {
		wrapped = append(wrapped, any(msg).(Msg))
	}

	return wrapped
}
