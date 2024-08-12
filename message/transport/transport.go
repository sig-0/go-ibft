package transport

import "github.com/sig-0/go-ibft/message"

type MulticastFn[M message.IBFTMessage] func(M)

func (f MulticastFn[M]) Multicast(msg M) {
	f(msg)
}

type transport struct {
	proposal    MulticastFn[*message.MsgProposal]
	prepare     MulticastFn[*message.MsgPrepare]
	commit      MulticastFn[*message.MsgCommit]
	roundChange MulticastFn[*message.MsgRoundChange]
}

// NewTransport returns a message.Transport object based on the provided message callbacks
func NewTransport(
	proposal MulticastFn[*message.MsgProposal],
	prepare MulticastFn[*message.MsgPrepare],
	commit MulticastFn[*message.MsgCommit],
	roundChange MulticastFn[*message.MsgRoundChange],
) message.Transport {
	return transport{
		proposal:    proposal,
		prepare:     prepare,
		commit:      commit,
		roundChange: roundChange,
	}
}

func (t transport) MulticastProposal(msg *message.MsgProposal) {
	t.proposal.Multicast(msg)
}

func (t transport) MulticastPrepare(msg *message.MsgPrepare) {
	t.prepare.Multicast(msg)
}

func (t transport) MulticastCommit(msg *message.MsgCommit) {
	t.commit.Multicast(msg)
}

func (t transport) MulticastRoundChange(msg *message.MsgRoundChange) {
	t.roundChange.Multicast(msg)
}
