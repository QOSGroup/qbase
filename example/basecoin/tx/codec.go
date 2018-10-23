package tx

import (
	"github.com/tendermint/go-amino"
)

func RegisterCodec(cdc *amino.Codec) {
	cdc.RegisterConcrete(&SendTx{}, "basecoin/SendTx", nil)
}
