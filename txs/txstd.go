package txs

import (
	"fmt"
	"github.com/QOSGroup/qbase/context"
	"github.com/QOSGroup/qbase/types"
	"github.com/tendermint/tendermint/crypto"
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
	ret = append(ret, []byte(tx.ChainID)...)
	ret = append(ret, types.Int2Byte(tx.MaxGas.Int64())...)

	return ret
}

//签名：每个签名者外部调用此方法
func (tx *TxStd) SignTx(privkey crypto.PrivKey, nonce int64) bool {
	if tx.ITx == nil {
		return false
	}

	sigdata := append(tx.GetSignData(), types.Int2Byte(nonce)...)
	prvdata, err := privkey.Sign(sigdata)
	if err != nil {
		return false
	}
	sig := Signature{privkey.PubKey(),
		prvdata,
		int64(nonce)}
	tx.Signature = append(tx.Signature, sig)

	return true
}

//构建结构体
//调用 NewTxStd后，需调用TxStd.SignTx填充TxStd.Signature(每个TxStd.Signer())
func NewTxStd(itx ITx, cid string, mgas types.BigInt) (rTx *TxStd) {
	rTx = new(TxStd)
	rTx.ITx = itx
	rTx.ChainID = cid
	rTx.MaxGas = mgas

	return
}

//函数：Signature结构转化为 []byte
func Sig2Byte(sgn Signature) (ret []byte) {
	if sgn.Pubkey == nil {
		//_, file, line, _ := runtime.Caller(1)
		//fmt.Printf("[%s](%d): pubkey is nil", file, line)
		return nil
	}
	ret = append(ret, sgn.Pubkey.Bytes()...)
	ret = append(ret, sgn.Signature...)
	ret = append(ret, types.Int2Byte(sgn.Nonce)...)

	return
}

func CopyTxStd(dst *TxStd, src *TxStd) bool {
	if src == nil || dst == nil {
		return false
	}

	dst.MaxGas = src.MaxGas
	dst.ChainID = src.ChainID
	for _, sg := range src.Signature {
		sig := Signature{sg.Pubkey, sg.Signature, sg.Nonce}
		dst.Signature = append(dst.Signature, sig)
	}
	dst.ITx = src.ITx

	return true
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
