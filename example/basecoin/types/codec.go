package types

import (
	"github.com/tendermint/go-amino"
)

func RegisterCodec(cdc *amino.Codec) {
	cdc.RegisterConcrete(&AppAccount{}, "basecoin/AppAccount", nil)
}
