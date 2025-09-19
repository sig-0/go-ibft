package sequencer

import (
	"github.com/sig-0/go-ibft/message"
)

// Sequence is a collection of consensus artifacts obtained by Sequencer during Finalize
type Sequence struct {
	// proposal that's being voted on
	proposal *message.Proposal

	// proposal that passed the PREPARE phase
	latestPB *message.ProposedBlock

	// proof that PREPARE was successful
	latestPC *message.PreparedCertificate

	// proof that ROUND CHANGE happened
	rcc *message.RoundChangeCertificate

	// proof that the proposal passed COMMIT phase
	seals []CommitSeal

	// currently running sequence
	sequence uint64

	// currently running round
	round uint64
}
