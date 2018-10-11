package baseabci

import (
	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/txs"
	go_amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto/encoding/amino"
)

var cdc = go_amino.NewCodec()

func init() {
	cryptoAmino.RegisterAmino(cdc)
	RegisterCodec(cdc)
}

func RegisterCodec(cdc *go_amino.Codec) {
	txs.RegisterCodec(cdc)
	account.RegisterCodec(cdc)
}
