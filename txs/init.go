package txs

import (
	"github.com/QOSGroup/qbase/types"
	go_amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto/encoding/amino"
)

var cdc = go_amino.NewCodec()

func init() {
	cryptoAmino.RegisterAmino(cdc)
	RegisterCodec(cdc)
}

func RegisterCodec(cdc *go_amino.Codec) {
	cdc.RegisterConcrete(&QcpTxResult{}, "qbase/txs/qcpresult", nil)
	cdc.RegisterConcrete(&Signature{}, "qbase/txs/signature", nil)
	cdc.RegisterConcrete(&TxStd{}, "qbase/txs/stdtx", nil)
	cdc.RegisterConcrete(&TxQcp{}, "qbase/txs/qcptx", nil)
	cdc.RegisterInterface((*ITx)(nil), nil)
	cdc.RegisterInterface((*types.Tx)(nil), nil)
}
