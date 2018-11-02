package tx

import (
	"bytes"
	"errors"

	"github.com/QOSGroup/qbase/baseabci"
	"github.com/QOSGroup/qbase/context"
	"github.com/QOSGroup/qbase/example/basecoin/types"
	"github.com/QOSGroup/qbase/txs"
	btypes "github.com/QOSGroup/qbase/types"
)

type SendTx struct {
	From btypes.Address  `json:"from"`
	To   btypes.Address  `json:"to"`
	Coin btypes.BaseCoin `json:"coin"`
}

var _ txs.ITx = (*SendTx)(nil)

func NewSendTx(from btypes.Address, to btypes.Address, coin btypes.BaseCoin) SendTx {
	return SendTx{From: from, To: to, Coin: coin}
}

func (tx *SendTx) ValidateData(ctx context.Context) error {
	if len(tx.From) == 0 || len(tx.To) == 0 || btypes.NewInt(0).GT(tx.Coin.Amount) {
		return errors.New("SendTx ValidateData error")
	}
	return nil
}

func (tx *SendTx) Exec(ctx context.Context) (result btypes.Result, crossTxQcps *txs.TxQcp) {
	result = btypes.Result{
		Code: btypes.ABCICodeOK,
	}
	// 查询发送方账户信息
	mapper := baseabci.GetAccountMapper(ctx)
	fromAcc := mapper.GetAccount(tx.From).(*types.AppAccount)
	if fromAcc.AccountAddress == nil {
		result.Code = btypes.ABCICodeType(btypes.CodeInternal)
		return
	}
	// 校验发送金额
	exists := false
	for _, c := range fromAcc.Coins {
		if c.Name == tx.Coin.Name {
			exists = true
			if c.Amount.LT(tx.Coin.Amount) {
				result.Code = btypes.ABCICodeType(btypes.CodeInternal)
				return
			}
		}
	}
	if !exists {
		result.Code = btypes.ABCICodeType(btypes.CodeInternal)
		return
	}

	// 查询接收方账户信息
	toAcc := mapper.GetAccount(tx.To)
	if toAcc == nil {
		toAcc = mapper.NewAccountWithAddress(tx.To).(*types.AppAccount)
	}
	toAccount := toAcc.(*types.AppAccount)
	// 更新账户状态
	for i, c := range fromAcc.Coins {
		if c.Name == tx.Coin.Name {
			fromAcc.Coins[i].Amount = c.Amount.Add(tx.Coin.Amount.Neg())
		}
	}
	mapper.SetAccount(fromAcc)
	exists = false
	for i, c := range toAccount.Coins {
		if c.Name == tx.Coin.Name {
			exists = true
			toAccount.Coins[i].Amount = c.Amount.Add(tx.Coin.Amount)
		}
	}
	if !exists {
		toAccount.Coins = append(toAccount.Coins, &(tx.Coin))
	}
	mapper.SetAccount(toAccount)
	return
}

func (tx *SendTx) GetSigner() []btypes.Address {
	return []btypes.Address{tx.From}
}

func (tx *SendTx) CalcGas() btypes.BigInt {
	return btypes.ZeroInt()
}

func (tx *SendTx) GetGasPayer() btypes.Address {
	return tx.From
}

func (tx *SendTx) GetSignData() []byte {
	var buf bytes.Buffer
	buf.Write(tx.From)
	buf.Write(tx.To)
	buf.Write([]byte(tx.Coin.String()))
	return buf.Bytes()
}
