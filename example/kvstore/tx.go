package kvstore

import (
	"errors"

	"github.com/QOSGroup/qbase/baseabci"
	"github.com/QOSGroup/qbase/context"
	"github.com/QOSGroup/qbase/txs"
	"github.com/QOSGroup/qbase/types"
)

type KvstoreTx struct {
	Key   []byte
	Value []byte
	Bytes []byte
}

var _ txs.ITx = (*KvstoreTx)(nil)

func NewKvstoreTx(key []byte, value []byte) KvstoreTx {
	return KvstoreTx{
		Key:   key,
		Value: value,
	}
}

func (kv KvstoreTx) ValidateData(ctx context.Context) error {
	if len(kv.Key) < 0 {
		return errors.New("key is empty")
	}
	return nil
}

func (kv KvstoreTx) Exec(ctx context.Context) (result types.Result, crossTxQcps *txs.TxQcp) {
	logger := ctx.Logger()

	//获取注册的mapper：
	kvMapper := ctx.Mapper(KVMapperName).(*KvMapper)
	//以下两个为qbase内置的mapper
	//QcpMapper: 跨链相关的操作
	//AccountMapper: 账户相关的操作

	accMapper := baseabci.GetAccountMapper(ctx)
	qcpMapper := baseabci.GetQcpMapper(ctx)

	logger.Info("kvMapper", kvMapper)
	logger.Info("qcpMapper", qcpMapper)
	logger.Info("accMapper", accMapper)

	key := string(kv.Key)
	value := kvMapper.GetKey(key)
	logger.Info("origin is: ", key, "=", value)

	kvMapper.SaveKV(key, string(kv.Value))
	value = kvMapper.GetKey(key)
	logger.Info("after is: ", key, value)

	return
}

func (kv KvstoreTx) GetSigner() []types.Address {
	return nil
}

func (kv KvstoreTx) CalcGas() types.BigInt {
	return types.ZeroInt()
}

func (kv KvstoreTx) GetGasPayer() types.Address {
	return types.Address{}
}

func (kv KvstoreTx) GetSignData() []byte {
	signData := make([]byte, len(kv.Key)+len(kv.Value)+len(kv.Bytes))
	signData = append(signData, kv.Key...)
	signData = append(signData, kv.Value...)
	signData = append(signData, kv.Bytes...)

	return signData
}
