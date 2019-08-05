package txs

import (
	"fmt"

	"github.com/QOSGroup/qbase/context"
	"github.com/QOSGroup/qbase/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/crypto"
)

// 功能：抽象具体的Tx结构体
type ITx interface {
	ValidateData(ctx context.Context) error //检测

	//执行业务逻辑,
	// crossTxQcp: 需要进行跨链处理的TxQcp。
	// 业务端实现中crossTxQcp只需包含`to` 和 `txStd`
	Exec(ctx context.Context) (result types.Result, crossTxQcp *TxQcp)
	GetSigner() []types.Address //签名者
	CalcGas() types.BigInt      //计算gas
	GetGasPayer() types.Address //gas付费人
	GetSignData() []byte        //获取签名字段
}

// 标准Tx结构体
type TxStd struct {
	ITxs      []ITx        `json:"itx"`      //ITx接口，将被具体Tx结构实例化
	Signature []Signature  `json:"sigature"` //签名数组
	ChainID   string       `json:"chainid"`  //ChainID: 执行ITx.exec方法的链ID
	MaxGas    types.BigInt `json:"maxgas"`   //Gas消耗的最大值
}

var _ types.Tx = (*TxStd)(nil)

// 签名结构体
type Signature struct {
	Pubkey    crypto.PubKey `json:"pubkey"`    //可选
	Signature []byte        `json:"signature"` //签名内容
	Nonce     int64         `json:"nonce"`     //nonce的值
}

// Type: just for implements types.Tx
func (tx *TxStd) Type() string {
	return "txstd"
}

func (tx *TxStd) GetSigners() []types.Address {

	if len(tx.ITxs) == 0 {
		panic("ITx shouldn't be nil in TxStd.GetSigners()")
	}

	var originSigners []types.Address
	for _, itx := range tx.ITxs {
		originSigners = append(originSigners, itx.GetSigner()...)
	}

	if len(originSigners) <= 1 {
		return originSigners
	}

	signers := make([]types.Address, 0, len(originSigners))
	m := make(map[string]struct{})

	for _, signer := range originSigners {
		if _, ok := m[signer.String()]; !ok {
			m[signer.String()] = struct{}{}
			signers = append(signers, signer)
		}
	}
	return signers
}

//BuildSignatureBytes 生成待签名字节切片.
//nonce: account nonce + 1
//currentChainID: 当前链chainID
func (tx *TxStd) BuildSignatureBytes(nonce int64, fromChainID string) []byte {
	bz := tx.getSignData()
	bz = append(bz, types.Int2Byte(nonce)...)
	if fromChainID != "" && fromChainID != tx.ChainID {
		bz = append(bz, []byte(fromChainID)...)
	} else {
		bz = append(bz, []byte(tx.ChainID)...)
	}
	return bz
}

// 获取TxStd中需要参与签名的内容
func (tx *TxStd) getSignData() []byte {
	if len(tx.ITxs) == 0 {
		panic("ITx shouldn't be nil in TxStd.GetSignData()")
	}

	var ret []byte
	for _, itx := range tx.ITxs {
		ret = append(ret, itx.GetSignData()...)
	}
	ret = append(ret, []byte(tx.ChainID)...)
	ret = append(ret, types.Int2Byte(tx.MaxGas.Int64())...)

	return ret
}

// 签名：每个签名者外部调用此方法
// 当tx不包含在跨链交易中时,fromChainID为 ""
func (tx *TxStd) SignTx(privkey crypto.PrivKey, nonce int64, fromChainID, toChainID string) (signedbyte []byte, err error) {
	if len(tx.ITxs) == 0 {
		return nil, errors.New("Signature txstd err(itx is nil)")
	}

	if tx.ChainID != toChainID {
		return nil, errors.New("toChainID not match txStd's chainID")
	}

	bz := tx.BuildSignatureBytes(nonce, fromChainID)
	signedbyte, err = privkey.Sign(bz)
	if err != nil {
		return nil, err
	}

	return
}

// 构建结构体
// 调用 NewTxStd后，需调用TxStd.SignTx填充TxStd.Signature(每个TxStd.Signer())
func NewTxStd(itx ITx, cid string, mgas types.BigInt) (rTx *TxStd) {
	rTx = NewTxsStd(cid, mgas, itx)

	return
}

func NewTxsStd(cid string, mgas types.BigInt, itx ...ITx) (rTx *TxStd) {
	rTx = &TxStd{
		itx,
		[]Signature{},
		cid,
		mgas,
	}

	return
}

// 函数：Signature结构转化为 []byte
func Sig2Byte(sgn Signature) (ret []byte) {
	if sgn.Pubkey == nil {
		return nil
	}
	ret = append(ret, sgn.Pubkey.Bytes()...)
	ret = append(ret, sgn.Signature...)
	ret = append(ret, types.Int2Byte(sgn.Nonce)...)

	return
}

//ValidateBasicData  对txStd进行基础的数据校验
func (tx *TxStd) ValidateBasicData(ctx context.Context, isCheckTx bool, currentChaindID string) (err types.Error) {
	if len(tx.ITxs) == 0 {
		return types.ErrInternal("TxStd's ITx is nil")
	}

	//开启cache执行ITx.ValidateData，在ITx.ValidateData中做数据保存操作将被忽略
	newCtx, _ := ctx.CacheContext()
	for _, itx := range tx.ITxs {
		itxErr := itx.ValidateData(newCtx)
		if itxErr != nil {
			return types.ErrInternal(fmt.Sprintf("TxStd's ITx ValidateData error:  %s", itxErr.Error()))
		}
	}

	if tx.ChainID == "" {
		return types.ErrInternal("TxStd's ChainID is empty")
	}

	if tx.ChainID != currentChaindID {
		return types.ErrInternal(fmt.Sprintf("chainId not match. expect: %s , actual: %s", currentChaindID, tx.ChainID))
	}

	if tx.MaxGas.LT(types.ZeroInt()) {
		return types.ErrInternal("TxStd's MaxGas is less than zero")
	}

	return
}
