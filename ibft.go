package go_ibft

import (
	"context"

	"github.com/madz-lab/go-ibft/message/types"
)

type (
	Signer     = types.Signer
	SigRecover = types.SigRecover
	Feed       = types.MsgFeed

	Transport interface {
		Multicast(types.Msg)
	}

	Quorum interface {
		HasQuorum(uint64, []types.Msg) bool
	}

	Keccak interface {
		Hash([]byte) []byte
	}

	Validator interface {
		Signer
		ID() []byte
		BuildBlock(uint64) []byte
	}

	Verifier interface {
		IsProposer(id []byte, sequence uint64, round uint64) bool
		IsValidator(id []byte, sequence uint64) bool
		IsValidBlock(block []byte, sequence uint64) bool
	}
)

type Context struct {
	context.Context
}

func NewIBFTContext(ctx context.Context) Context {
	return Context{ctx}
}

func (c Context) WithCancel() (Context, func()) {
	subCtx, cancelFn := context.WithCancel(c)
	return Context{subCtx}, cancelFn
}

func (c Context) WithTransport(t Transport) Context {
	return Context{context.WithValue(c, "transport", t)}
}

func (c Context) Transport() Transport {
	return c.Value("transport").(Transport)
}

func (c Context) WithFeed(f Feed) Context {
	return Context{context.WithValue(c, "feed", f)}
}

func (c Context) Feed() Feed {
	return c.Value("feed").(Feed)
}

func (c Context) WithQuorum(q Quorum) Context {
	return Context{context.WithValue(c, "quorum", q)}
}

func (c Context) Quorum() Quorum {
	return c.Value("quorum").(Quorum)
}

func (c Context) WithKeccak(k Keccak) Context {
	return Context{context.WithValue(c, "keccak", k)}
}

func (c Context) Keccak() Keccak {
	return c.Value("keccak").(Keccak)
}

func (c Context) WithSigRecover(s SigRecover) Context {
	return Context{context.WithValue(c, "sig_recover", s)}
}

func (c Context) SigRecover() SigRecover {
	return c.Value("sig_recover").(SigRecover)
}
