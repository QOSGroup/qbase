package txs

import (
	"fmt"
	"github.com/QOSGroup/qbase/context"
	"github.com/QOSGroup/qbase/types"
	"github.com/tendermint/tendermint/crypto"
	"log"
)

//功能：抽象具体的Tx结构体
type ITx interface {
	ValidateData() bool                                                 //检测
	Exec(ctx context.Context) (result types.Result, crossTxQcps *TxQcp) //执行, crossTxQcps: 需要跨链处理的TxQcp
	GetSigner() []types.Address                                         //签名者
	CalcGas() types.BigInt                                              //计算gas
	GetGasPayer() types.Address                                         //gas付费人
	GetSignData() []byte                                                //获取签名字段
}

//标准Tx结构体
type TxStd struct {
	ITx       ITx          `json:"itx"`      //ITx接口，将被具体Tx结构实例化
	Signature []Signature  `json:"sigature"` //签名数组
	ChainID   string       `json:"chainid"`  //ChainID
	MaxGas    types.BigInt `json:"maxgas"`   //Gas消耗的最大值
}

//签名结构体
type Signature struct {
	Pubkey    crypto.PubKey `json:"pubkey"`    //可选
	Signature []byte        `json:"signature"` //签名内容
	Nonce     int64         `json:"nonce"`     //nonce的值
}

//Type: just for implements types.Tx
func (tx *TxStd) Type() string {
	return "txstd"
}

//将需要签名的字段拼接成 []byte
func (tx *TxStd) GetSignData() []byte {
	if tx.ITx == nil {
		panic("ITx shouldn't be nil in TxStd.GetSignData()")
		return nil
	}

	ret := tx.ITx.GetSignData()
	for _, sgn := range tx.Signature {
		ret = Sig2Byte(sgn)
	}

	ret = append(ret, []byte(tx.ChainID)...)
	ret = append(ret, types.Int2Byte(tx.MaxGas.Int64())...)
	return ret
}

//构建结构体
func NewTxStd(itx ITx, cid string, mgas types.BigInt) (rTx *TxStd) {
	rTx = new(TxStd)
	rTx.ITx = itx
	rTx.ChainID = cid
	rTx.MaxGas = mgas
	return
}

//函数：Signature结构转化为 []byte
func Sig2Byte(sgn Signature) []byte {
	var ret = []byte{}
	ret = append(ret, []byte(sgn.Pubkey.Bytes())...)
	ret = append(ret, sgn.Signature...)
	ret = append(ret, types.Int2Byte(sgn.Nonce)...)
	return ret
}

func CopyTxStd(dst *TxStd, src *TxStd) {
	if src == nil || dst == nil {
		log.Panic("Input struct can't be nil")
		return
	}

	dst.MaxGas = src.MaxGas
	dst.ChainID = src.ChainID
	dst.Signature = src.Signature
	dst.ITx = src.ITx
}

//ValidateBasicData  对txStd进行基础的数据校验
//tx.ITx == QcpTxResult时 不校验签名相关信息
func (tx *TxStd) ValidateBasicData(isCheckTx bool, currentChaindID string) (err types.Error) {
	if tx.ITx == nil {
		return types.ErrInternal("no itx in txStd")
	}

	if !tx.ITx.ValidateData() {
		return types.ErrInternal("invaild ITx data in txStd")
	}

	if tx.ChainID == "" {
		return types.ErrInternal("no chainId in txStd")
	}

	if tx.ChainID != currentChaindID {
		return types.ErrInternal(fmt.Sprintf("chainId not match. expect: %s , actual: %s", currentChaindID, tx.ChainID))
	}

	if tx.MaxGas.LT(types.ZeroInt()) {
		return types.ErrInternal("invaild max gas in txStd")
	}

	_, ok := tx.ITx.(*QcpTxResult)
	if !ok {

		singers := tx.ITx.GetSigner()
		if len(singers) == 0 {
			return
		}

		sigs := tx.Signature
		if len(sigs) == 0 {
			return types.ErrUnauthorized("no signatures")
		}

		if len(sigs) != len(singers) {
			return types.ErrUnauthorized("signatures and signers not match")
		}
	}

	return
}
