package tx

import (
	"bytes"
	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/context"
	bctypes "github.com/QOSGroup/qbase/example/basecoin/types"
	"github.com/QOSGroup/qbase/txs"
	"github.com/QOSGroup/qbase/types"
)

type SendTx struct {
	From      types.Address `json:"from"`
	To        types.Address `json:"to"`
	Coin      bctypes.Coin  `json:"coin"`
}

func NewSendTx(from types.Address, to types.Address, coin bctypes.Coin) SendTx {
	return SendTx{From: from, To: to, Coin: coin}
}

func (tx *SendTx) ValidateData() bool {
	if len(tx.From) == 0 || len(tx.To) == 0 || !tx.Coin.IsNotNegative() {
		return false
	}
	return true
}

func (tx *SendTx) Exec(ctx context.Context) (result types.Result, crossTxQcps *txs.TxQcp) {
	result = types.Result{
		Code: types.ABCICodeOK,
	}
	//查询发送方账户信息，校验发送金额
	mapper := ctx.Mapper(account.AccountMapperName).(*account.AccountMapper)
	fromAcc := mapper.GetAccount(tx.From).(*bctypes.AppAccount)
	if fromAcc.AccountAddress == nil || fromAcc.Coins.AmountOf(tx.Coin.Name).LT(tx.Coin.Amount) {
		result.Code = types.ABCICodeType(types.CodeInternal)
		return
	}
	//查询接收方账户信息
	toAcc := mapper.GetAccount(tx.To)
	if toAcc == nil {
		toAcc = mapper.NewAccountWithAddress(tx.To).(*bctypes.AppAccount)
	}
	toAccount := toAcc.(*bctypes.AppAccount)
	//更新账户状态
	fromAcc.Coins = fromAcc.Coins.MinusSingle(tx.Coin)
	mapper.SetAccount(fromAcc)
	toAccount.Coins = toAccount.Coins.PlusSingle(tx.Coin)
	mapper.SetAccount(toAccount)
	return
}

func (tx *SendTx) GetSigner() []types.Address {
	return []types.Address{tx.From}
}

func (tx *SendTx) CalcGas() types.BigInt {
	return types.ZeroInt()
}

func (tx *SendTx) GetGasPayer() types.Address {
	return tx.From
}

func (tx *SendTx) GetSignData() []byte {
	var buf bytes.Buffer
	buf.Write(tx.From)
	buf.Write(tx.To)
	buf.Write([]byte(tx.Coin.String()))
	return buf.Bytes()
}
