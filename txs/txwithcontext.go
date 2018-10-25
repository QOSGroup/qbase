package txs

import (
	"github.com/QOSGroup/qbase/context"
)

type TxWithContext struct {
	ctx context.Context
}

func NewTxWithContext(ctx context.Context) *TxWithContext {
	return &TxWithContext{
		ctx: ctx,
	}
}

func (bct *TxWithContext) CurrentContext() context.Context {
	return bct.ctx
}
