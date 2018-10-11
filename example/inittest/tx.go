package inittest

import (
	"github.com/QOSGroup/qbase/context"
	"github.com/QOSGroup/qbase/txs"
	"github.com/QOSGroup/qbase/types"
)

type InitTestTx struct {
	Key   []byte
	Value []byte
	Bytes []byte
}

func NewInitTestTx(key []byte, value []byte) InitTestTx {
	return InitTestTx{
		Key:   key,
		Value: value,
	}
}

func (kv InitTestTx) ValidateData() bool {
	if len(kv.Key) < 0 {
		return false
	}
	return true
}

func (kv InitTestTx) Exec(ctx context.Context) (result types.Result, crossTxQcps *txs.TxQcp) {

	txMapper := ctx.Mapper(TX_MAPPER_NAME).(*TxMapper)

	key := string(kv.Key)
	val := string(kv.Value)

	txMapper.SaveKV(key, val)
	result = types.Result{
		Code: types.ABCICodeOK,
		Data: kv.Value,
	}
	return
}

func (kv InitTestTx) GetSigner() []types.Address {
	return nil
}

func (kv InitTestTx) CalcGas() types.BigInt {
	return types.ZeroInt()
}

func (kv InitTestTx) GetGasPayer() types.Address {
	return types.Address{}
}

func (kv InitTestTx) GetSignData() []byte {
	signData := make([]byte, len(kv.Key)+len(kv.Value)+len(kv.Bytes))
	signData = append(signData, kv.Key...)
	signData = append(signData, kv.Value...)
	signData = append(signData, kv.Bytes...)

	return signData
}
