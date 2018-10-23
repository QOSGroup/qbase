package kvstore

import (
	"github.com/QOSGroup/qbase/baseabci"
	go_amino "github.com/tendermint/go-amino"
)

func MakeKVStoreCodec() *go_amino.Codec {
	cdc := baseabci.MakeQBaseCodec()
	RegisterCodec(cdc)
	return cdc
}

func RegisterCodec(cdc *go_amino.Codec) {
	cdc.RegisterConcrete(&KvstoreTx{}, "kvstore/main/kvstoretx", nil)
}
