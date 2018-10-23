package app

import (
	"github.com/QOSGroup/qbase/baseabci"
	"github.com/QOSGroup/qbase/example/basecoin/tx"
	"github.com/QOSGroup/qbase/example/basecoin/types"
	"github.com/tendermint/go-amino"
)

func MakeCodec() *amino.Codec {
	cdc := baseabci.MakeQBaseCodec()
	RegisterCodec(cdc)
	return cdc
}

func RegisterCodec(cdc *amino.Codec) {
	cdc.RegisterConcrete(&types.AppAccount{}, "basecoin/AppAccount", nil)
	cdc.RegisterConcrete(&tx.SendTx{}, "basecoin/SendTx", nil)
}
