package baseabci

import (
	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/txs"
	go_amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto/encoding/amino"
)

func MakeQBaseCodec() *go_amino.Codec {

	var cdc = go_amino.NewCodec()
	cryptoAmino.RegisterAmino(cdc)
	txs.RegisterCodec(cdc)
	account.RegisterCodec(cdc)

	return cdc
}
