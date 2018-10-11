package inittest

import (
	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/types"
	"github.com/tendermint/tendermint/crypto"
)

// QOSAccount定义基本账户之上的QOS和QSC
type QOSAccount struct {
	BaseAccount account.BaseAccount `json:"base_account"` // inherits BaseAccount
	Qos         types.BigInt        `json:"qos"`          // coins in public chain
	QscList     []*QSC              `json:"qsc"`          // varied QSCs
}

func ProtoQOSAccount() account.Account {
	return &QOSAccount{}
}

func (account *QOSAccount) GetAddress() types.Address {
	return account.BaseAccount.GetAddress()
}

func (account *QOSAccount) SetAddress(addr types.Address) error {
	return account.BaseAccount.SetAddress(addr)
}

func (account *QOSAccount) GetPubicKey() crypto.PubKey {
	return account.BaseAccount.GetPubicKey()
}

func (account *QOSAccount) SetPublicKey(pubKey crypto.PubKey) error {
	return account.BaseAccount.SetPublicKey(pubKey)
}

func (account *QOSAccount) GetNonce() uint64 {
	return account.BaseAccount.GetNonce()
}

func (account *QOSAccount) SetNonce(nonce uint64) error {
	return account.BaseAccount.SetNonce(nonce)
}

// 获得账户QOS的数量
func (accnt *QOSAccount) GetQOS() types.BigInt {
	return accnt.Qos
}

// 设置账户QOS的数量
func (accnt *QOSAccount) SetQOS(amount types.BigInt) error {
	accnt.Qos = amount
	return nil
}

// 获取账户的名为QSCName的币的数量
func (accnt *QOSAccount) getQSC(QSCName string) *QSC {
	for _, qsc := range accnt.QscList {
		if qsc.GetName() == QSCName {
			return qsc
		}
	}
	return nil
}

// 设置账户的名为QSCName的币
func (accnt *QOSAccount) setQSC(newQSC *QSC) error {
	for _, qsc := range accnt.QscList {
		if qsc.GetName() == newQSC.GetName() {
			qsc.SetAmount(newQSC.GetAmount())
			return nil
		}
	}
	accnt.QscList = append(accnt.QscList, newQSC)
	return nil
}

// 删除账户中名为QSCName的币
func (accnt *QOSAccount) removeQSCByName(QSCName string) error {
	for i, qsc := range accnt.QscList {
		if qsc.GetName() == QSCName {
			if i == len(accnt.QscList)-1 {
				accnt.QscList = accnt.QscList[:i]
				return nil
			}
			accnt.QscList = append(accnt.QscList[:i], accnt.QscList[i+1:]...)
			return nil
		}
	}
	return types.ErrInvalidCoins(types.CodeToDefaultMsg(types.CodeInvalidCoins))
}
