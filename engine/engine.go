package engine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/store"
	"github.com/madz-lab/go-ibft/message/types"
	"github.com/madz-lab/go-ibft/sequencer"
)

var (
	ErrInvalidConfig  = errors.New("invalid engine config")
	ErrInvalidMessage = errors.New("invalid ibft message")
)

type Config struct {
	TransportMsgProposal    ibft.Transport[*types.MsgProposal]
	TransportMsgPrepare     ibft.Transport[*types.MsgPrepare]
	TransportMsgCommit      ibft.Transport[*types.MsgCommit]
	TransportMsgRoundChange ibft.Transport[*types.MsgRoundChange]
	Quorum                  ibft.Quorum
	Keccak                  ibft.Keccak
	Round0Duration          time.Duration
}

func (cfg Config) IsValid() error {
	if cfg.TransportMsgProposal == nil ||
		cfg.TransportMsgPrepare == nil ||
		cfg.TransportMsgRoundChange == nil ||
		cfg.TransportMsgCommit == nil {
		return fmt.Errorf("%w: missing message transport", ErrInvalidConfig)
	}

	if cfg.Quorum == nil {
		return fmt.Errorf("%w: missing quorum", ErrInvalidConfig)
	}

	if cfg.Keccak == nil {
		return fmt.Errorf("%w: missing keccak", ErrInvalidConfig)
	}

	if cfg.Round0Duration == 0 {
		return fmt.Errorf("%w: round zero duration cannot be 0", ErrInvalidConfig)
	}

	return nil
}

type Engine struct {
	*sequencer.Sequencer

	messages *store.MessageStore
	cfg      Config
}

func NewEngine(validator ibft.Validator, cfg Config) Engine {
	return Engine{
		Sequencer: sequencer.New(validator, cfg.Round0Duration),
		messages:  store.NewMessageStore(),
		cfg:       cfg,
	}
}

func (e Engine) AddMessage(msg types.Message) error {
	if err := msg.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidMessage, err)
	}

	if !e.Validator.IsValidSignature(
		msg.GetSender(),
		e.cfg.Keccak.Hash(msg.Payload()),
		msg.GetSignature(),
	) {
		return fmt.Errorf("%w: invalid signature", ErrInvalidMessage)
	}

	switch msg := msg.(type) {
	case *types.MsgProposal:
		e.messages.ProposalMessages.AddMessage(msg)
	case *types.MsgPrepare:
		e.messages.PrepareMessages.AddMessage(msg)
	case *types.MsgCommit:
		e.messages.CommitMessages.AddMessage(msg)
	case *types.MsgRoundChange:
		e.messages.RoundChangeMessages.AddMessage(msg)
	}

	return nil
}

type SequenceResult struct {
	Sequence         uint64
	SequenceProposal *types.FinalizedProposal
}

func (e Engine) FinalizeSequence(ctx context.Context, sequence uint64) SequenceResult {
	defer func() {
		e.messages.ProposalMessages.Clear()
		e.messages.PrepareMessages.Clear()
		e.messages.CommitMessages.Clear()
		e.messages.RoundChangeMessages.Clear()
	}()

	c := sequencer.NewContext(ctx,
		sequencer.WithQuorum(e.cfg.Quorum),
		sequencer.WithKeccak(e.cfg.Keccak),
		sequencer.WithMessageFeed(e.messages.Feed()),
		sequencer.WithMessageTransport(sequencer.MessageTransport{
			Proposal:    e.cfg.TransportMsgProposal,
			Prepare:     e.cfg.TransportMsgPrepare,
			Commit:      e.cfg.TransportMsgCommit,
			RoundChange: e.cfg.TransportMsgRoundChange,
		}),
	)

	return SequenceResult{
		Sequence:         sequence,
		SequenceProposal: e.Finalize(c, sequence),
	}
}
