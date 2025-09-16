package sequencer

type Verifier interface {
	// IsValidator checks if id is part of consensus for given sequence
	IsValidator(addr []byte, sequence uint64) bool

	// IsProposer asserts if id is the elected proposer for given sequence and round
	IsProposer(addr []byte, sequence, round uint64) bool

	// HasQuorum returns true if messages accumulate consensus for a particular sequence
	HasQuorum(addresses [][]byte, sequence uint64) bool

	// IsValidProposal checks if the provided proposal is valid for given sequence
	IsValidProposal(proposal []byte, sequence uint64) bool

	// IsValidSignature checks if the signature came from signer over some digest
	IsValidSignature(signer, digest, signature []byte) error
}
