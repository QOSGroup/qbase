package txs

import (
	"fmt"
	"github.com/QOSGroup/qbase/types"
	"github.com/tendermint/tendermint/crypto"
)

//功能：
type TxQcp struct {
	Payload     TxStd     `json:"payload"`     //TxStd结构
	From        string    `json:"from"`        //qscName
	To          string    `json:"to"`          //qosName
	Sequence    int64     `json:"sequence"`    //发送Sequence
	Sig         Signature `json:"sig"`         //签名
	BlockHeight int64     `json:"blockheight"` //Tx所在block高度
	TxIndx      int64     `json:"txindx"`      //Tx在block的位置
	IsResult    bool      `json:"isresult"`    //是否为Result
}

//Type: just for implements types.Tx
func (tx *TxQcp) Type() string {
	return "txqcp"
}

//------------------------------------
//package example;
//
//enum FOO { X = 17; };
//
//message Test {
//	required string label = 1;
//	optional int32 type = 2 [default=77];
//	repeated int64 reps = 3;
//	optional group OptionalGroup = 4 {
//	required string RequiredField = 5;
//	}
//}
//-----------------------------------

//功能：获取 TxQcp 的签名字段
//返回：字段拼接后的 []byte
func (tx *TxQcp) GetSigData() []byte {
	ret := tx.Payload.GetSignData()
	if ret == nil {
		fmt.Print("GetSigData() is nil in TxQcp")
		return nil
	}

	ret = append(ret, []byte(tx.From)...)
	ret = append(ret, []byte(tx.To)...)
	ret = append(ret, types.Int2Byte(tx.Sequence)...)
	ret = append(ret, types.Int2Byte(tx.BlockHeight)...)
	ret = append(ret, types.Int2Byte(tx.TxIndx)...)
	ret = append(ret, types.Bool2Byte(tx.IsResult)...)

	return ret
}

func (tx *TxQcp) SignTx(prvkey crypto.PrivKey) bool {
	data := tx.GetSigData()
	if data == nil {
		return false
	}
	prvdata, err := prvkey.Sign(data)
	if err != nil {
		return false
	}

	tx.Sig.Pubkey = prvkey.PubKey()
	tx.Sig.Nonce = 0 //nonce shouldn't be used in TxQcp.Sig, use TxQcp.Sequence.
	tx.Sig.Signature = prvdata

	return true
}

//构建TxQCP结构体
func NewTxQCP(payLoad *TxStd, from string, to string, seq int64,
	bkheigh int64, tidx int64, isResult bool) (rTx *TxQcp) {

	rTx = new(TxQcp)
	CopyTxStd(&rTx.Payload, payLoad)
	rTx.From = from
	rTx.To = to
	rTx.Sequence = seq
	rTx.BlockHeight = bkheigh
	rTx.TxIndx = tidx
	rTx.IsResult = isResult

	return
}

//ValidateBasicData 校验txQcp基础数据是否合法
func (tx *TxQcp) ValidateBasicData(isCheckTx bool, currentChaindID string) (err types.Error) {
	//1. From To Sequence Sig 不为空
	//2. to == current.chainId

	if tx.From == "" || tx.To == "" || tx.Sequence == 0 || tx.Sig.Signature == nil {
		return types.ErrInternal("txQcp's basic data is not valid")
	}

	if tx.To != currentChaindID {
		return types.ErrInternal("txQcp's To chainID is not match current chainID")
	}

	return
}
