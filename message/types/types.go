package types

type (
	// FinalizedSeal is proof that a validator committed to a specific proposal
	FinalizedSeal struct {
		From, CommitSeal []byte
	}

	// FinalizedProposal is a consensus verified proposal of some sequence
	FinalizedProposal struct {
		// proposal that was finalized
		Proposal []byte

		// seals of validators who committed to this proposal
		Seals []FinalizedSeal

		// round in which the proposal was finalized
		Round uint64
	}

	Subscription[M IBFTMessage] chan MsgNotification[M]

	// MsgNotification is received from the subscription to indicate a new message
	MsgNotification[M IBFTMessage] interface {
		// Unwrap returns all messages that fit the subscription
		Unwrap() []M
	}
)

type MsgNotificationFn[M IBFTMessage] func() []M

func (r MsgNotificationFn[M]) Unwrap() []M {
	return r()
}

func (rcc *RoundChangeCertificate) HighestRoundBlock() ([]byte, uint64) {
	roundsAndPreparedBlocks := make(map[uint64][]byte)

	for _, msg := range rcc.Messages {
		pb := msg.LatestPreparedProposedBlock
		pc := msg.LatestPreparedCertificate

		if pb == nil || pc == nil {
			continue
		}

		roundsAndPreparedBlocks[pc.ProposalMessage.Round()] = pb.Block
	}

	if len(roundsAndPreparedBlocks) == 0 {
		return nil, 0
	}

	var (
		highestRound      uint64
		highestRoundBlock []byte
	)

	for round, block := range roundsAndPreparedBlocks {
		if round >= highestRound {
			highestRound = round
			highestRoundBlock = block
		}
	}

	return highestRoundBlock, highestRound
}

func (rcc *RoundChangeCertificate) HighestRoundBlockHash() ([]byte, uint64) {
	roundsAndPreparedBlockHashes := make(map[uint64][]byte)

	for _, msg := range rcc.Messages {
		pc := msg.LatestPreparedCertificate
		if pc == nil {
			continue
		}

		roundsAndPreparedBlockHashes[pc.ProposalMessage.Round()] = pc.ProposalMessage.BlockHash
	}

	if len(roundsAndPreparedBlockHashes) == 0 {
		return nil, 0
	}

	var (
		highestRound          uint64
		highestRoundBlockHash []byte
	)

	for round, proposalHash := range roundsAndPreparedBlockHashes {
		if round >= highestRound {
			highestRound = round
			highestRoundBlockHash = proposalHash
		}
	}

	return highestRoundBlockHash, highestRound
}
