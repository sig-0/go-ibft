package sequencer

import "github.com/madz-lab/go-ibft/message/types"

type Quorum interface {
	HasQuorumPrepareMessages(...*types.MsgPrepare) bool
	HasQuorumCommitMessages(...*types.MsgCommit) bool
}
