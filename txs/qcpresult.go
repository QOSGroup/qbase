package txs

import (
	"github.com/QOSGroup/qbase/context"
	"github.com/QOSGroup/qbase/types"
	tcommon "github.com/tendermint/tendermint/libs/common"
)

// qos端对TxQcp的执行结果
type QcpTxResult struct {
	Code                int64            `json:"code"`                //执行结果
	Extends             []tcommon.KVPair `json:"extends"`             //结果附加值
	GasUsed             types.BigInt     `json:"gasused"`             //gas消耗值
	QcpOriginalSequence int64            `json:"qcporiginalsequence"` //此结果对应的TxQcp.Sequence
	Info                string           `json:"info"`                //结果信息
}

var _ ITx = (*QcpTxResult)(nil)

// 功能：检测结构体字段的合法性
// todo:QcpOriginalSequence 加入检测
func (tx *QcpTxResult) ValidateData(ctx context.Context) bool {
	if tx.Extends == nil || len(tx.Extends) == 0 || types.BigInt.LT(tx.GasUsed, types.ZeroInt()) {
		return false
	}

	return true
}

// 功能：tx执行
// 备注：用户根据tx.QcpOriginalSequence,需自行实现此接口
func (tx *QcpTxResult) Exec(ctx context.Context) (result types.Result, crossTxQcps *TxQcp) {
	result = ctx.TxQcpResultHandler()(ctx, tx)

	return
}

// 功能：获取签名者
// 备注：qos对QcpTxResult不做签名，故返回空
func (tx *QcpTxResult) GetSigner() []types.Address {
	return nil
}

// 功能：计算gas
// 备注：暂返回0，后期可根据实际情况调整
func (tx *QcpTxResult) CalcGas() types.BigInt {
	return types.ZeroInt()
}

// 功能：获取gas付费人
// 备注：返回空(因暂时gas为0，无人需要付gas)
func (tx *QcpTxResult) GetGasPayer() types.Address {
	return nil
}

// 获取签名字段
func (tx *QcpTxResult) GetSignData() []byte {
	ret := types.Int2Byte(tx.Code)
	ret = append(ret, Extends2Byte(tx.Extends)...)
	ret = append(ret, types.Int2Byte(tx.GasUsed.Int64())...)
	ret = append(ret, types.Int2Byte(tx.QcpOriginalSequence)...)
	ret = append(ret, []byte(tx.Info)...)

	return ret
}

// 功能：构建 QcpTxReasult 结构体
func NewQcpTxResult(code int64, ext *[]tcommon.KVPair, sequence int64, gasused types.BigInt, info string) (rTx *QcpTxResult) {
	rTx = &QcpTxResult{
		code,
		*ext,
		gasused,
		sequence,
		info,
	}
	return rTx
}

// 功能：将common.KVPair转化成[]byte
// todo: test（amino序列化及反序列化的正确性）
func Extends2Byte(ext []tcommon.KVPair) (ret []byte) {
	if ext == nil || len(ext) == 0 {
		return nil
	}

	for _, kv := range ext {
		ret = append(ret, kv.GetKey()...)
		ret = append(ret, kv.GetValue()...)
	}

	return ret
}
