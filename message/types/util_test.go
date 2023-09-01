package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsValidPreparedCertificate(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name    string
		view    *View
		pc      *PreparedCertificate
		isValid bool
	}{
		{
			name: "nil proposal message",
			pc: &PreparedCertificate{
				ProposalMessage: nil,
			},
		},

		{
			name: "nil prepare messages",
			pc: &PreparedCertificate{
				PrepareMessages: nil,
			},
		},

		{
			name: "invalid sequence in proposal message",
			view: &View{Sequence: 101, Round: 0},
			pc: &PreparedCertificate{
				ProposalMessage: &MsgProposal{
					View: &View{Sequence: 201, Round: 0},
				},
				PrepareMessages: []*MsgPrepare{},
			},
		},

		{
			name: "invalid round in proposal message",
			view: &View{Sequence: 101, Round: 0},
			pc: &PreparedCertificate{
				ProposalMessage: &MsgProposal{
					View: &View{Sequence: 101, Round: 0},
				},
				PrepareMessages: []*MsgPrepare{},
			},
		},

		{
			name: "prepare message sequence does not match proposal message sequence",
			view: &View{Sequence: 101, Round: 0},
			pc: &PreparedCertificate{
				ProposalMessage: &MsgProposal{
					View: &View{Sequence: 101, Round: 0},
				},
				PrepareMessages: []*MsgPrepare{
					{
						View: &View{Sequence: 201, Round: 0},
					},
				},
			},
		},

		{
			name: "prepare message round does not match proposal message round",
			view: &View{Sequence: 101, Round: 1},
			pc: &PreparedCertificate{
				ProposalMessage: &MsgProposal{
					View: &View{Sequence: 101, Round: 0},
				},
				PrepareMessages: []*MsgPrepare{
					{
						View: &View{Sequence: 101, Round: 1},
					},
				},
			},
		},

		{
			name: "prepare message block hash does not match proposal message block hash",
			view: &View{Sequence: 101, Round: 1},
			pc: &PreparedCertificate{
				ProposalMessage: &MsgProposal{
					View:         &View{Sequence: 101, Round: 0},
					ProposalHash: []byte("proposal hash"),
				},
				PrepareMessages: []*MsgPrepare{
					{
						View:         &View{Sequence: 101, Round: 0},
						ProposalHash: []byte("invalid proposal hash"),
					},
				},
			},
		},

		{
			name: "duplicate senders in prepared certificate",
			view: &View{Sequence: 101, Round: 1},
			pc: &PreparedCertificate{
				ProposalMessage: &MsgProposal{
					View:         &View{Sequence: 101, Round: 0},
					From:         []byte("proposer"),
					ProposalHash: []byte("proposal hash"),
				},
				PrepareMessages: []*MsgPrepare{
					{
						View:         &View{Sequence: 101, Round: 0},
						From:         []byte("proposer"),
						ProposalHash: []byte("proposal hash"),
					},
				},
			},
		},

		{
			name: "valid pc",
			view: &View{Sequence: 101, Round: 1},
			pc: &PreparedCertificate{
				ProposalMessage: &MsgProposal{
					View:         &View{Sequence: 101, Round: 0},
					From:         []byte("proposer"),
					ProposalHash: []byte("proposal hash"),
				},
				PrepareMessages: []*MsgPrepare{
					{
						View:         &View{Sequence: 101, Round: 0},
						From:         []byte("some validator"),
						ProposalHash: []byte("proposal hash"),
					},
				},
			},
			isValid: true,
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.isValid, tt.pc.IsValid(tt.view))
		})
	}
}
