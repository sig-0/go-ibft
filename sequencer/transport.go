package sequencer

import "github.com/madz-lab/go-ibft/message/types"

type Transport interface {
	MulticastProposal(*types.MsgProposal)
	MulticastPrepare(*types.MsgPrepare)
	MulticastCommit(*types.MsgCommit)
	MulticastRoundChange(*types.MsgRoundChange)
}

var DummyTransport = dummyTransport{}

type dummyTransport struct{}

func (t dummyTransport) MulticastProposal(_ *types.MsgProposal) {}

func (t dummyTransport) MulticastPrepare(_ *types.MsgPrepare) {}

func (t dummyTransport) MulticastCommit(_ *types.MsgCommit) {}

func (t dummyTransport) MulticastRoundChange(_ *types.MsgRoundChange) {}
