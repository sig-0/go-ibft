package sequencer

type Codec interface {
	// Keccak computes the Keccak hash of given input
	Keccak([]byte) []byte
	// RecoverFrom extracts the sender of the associated data and signature
	RecoverFrom(data, sig []byte) []byte
}
