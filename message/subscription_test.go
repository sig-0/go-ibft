package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Subscribe_Proposal(t *testing.T) {
	t.Parallel()

	t.Run("exact sequence and round", func(t *testing.T) {
		t.Parallel()

		s := NewStore()

		sub, cancelSub := s.ProposalMessages.Subscribe(101, 0, false)
		defer cancelSub()

		s.ProposalMessages.Add(&Proposal{
			Sequence:  101,
			Round:     0,
			Sender:    []byte("sender"),
			Signature: []byte("signature"),
			ProposedBlock: &ProposedBlock{
				Block: []byte("block"),
				Round: 0,
			},
			BlockHash: []byte("block hash"),
		})

		notification := <-sub
		mm := notification()[0]

		assert.EqualValues(t, 101, mm.Sequence)
		assert.EqualValues(t, 0, mm.Round)
	})

	t.Run("highest available round", func(t *testing.T) {
		t.Parallel()

		s := NewStore()

		sub, cancelSub := s.ProposalMessages.Subscribe(101, 0, true)
		defer cancelSub()

		s.ProposalMessages.Add(&Proposal{
			Sequence:  101,
			Round:     0,
			Sender:    []byte("sender"),
			Signature: []byte("signature"),
			ProposedBlock: &ProposedBlock{
				Block: []byte("block"),
				Round: 0,
			},
			BlockHash: []byte("block hash"),
		})
		s.ProposalMessages.Add(&Proposal{
			Sequence:  101,
			Round:     1,
			Sender:    []byte("sender"),
			Signature: []byte("signature"),
			ProposedBlock: &ProposedBlock{
				Block: []byte("block"),
				Round: 0,
			},
			BlockHash: []byte("block hash"),
		})
		s.ProposalMessages.Add(&Proposal{
			Sequence:  101,
			Round:     2,
			Sender:    []byte("sender"),
			Signature: []byte("signature"),
			ProposedBlock: &ProposedBlock{
				Block: []byte("block"),
				Round: 0,
			},
			BlockHash: []byte("block hash"),
		})

		notification := <-sub
		mm := notification()[0]

		assert.EqualValues(t, 2, mm.Round)
	})
}
