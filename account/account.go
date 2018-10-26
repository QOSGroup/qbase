package account

import (
	"github.com/QOSGroup/qbase/types"
	"github.com/tendermint/tendermint/crypto"
)

type Account interface {
	GetAddress() types.Address
	SetAddress(addr types.Address) error
	GetPubicKey() crypto.PubKey
	SetPublicKey(pubKey crypto.PubKey) error
	GetNonce() int64
	SetNonce(nonce int64) error
}

type BaseAccount struct {
	AccountAddress types.Address `json:"account_address"` // account address
	Publickey      crypto.PubKey `json:"public_key"`      // public key
	Nonce          int64         `json:"nonce"`           // identifies tx_status of an account
}

func ProtoBaseAccount() Account {
	return &BaseAccount{}
}

// getter for account address
func (accnt *BaseAccount) GetAddress() types.Address {
	return accnt.AccountAddress
}

// setter for account address
func (accnt *BaseAccount) SetAddress(addr types.Address) error {
	if len(addr) == 0 {
		return types.ErrInvalidAddress(types.CodeToDefaultMsg(types.CodeInvalidAddress))
	}
	accnt.AccountAddress = addr
	return nil
}

// getter for public key
func (accnt *BaseAccount) GetPubicKey() crypto.PubKey {
	return accnt.Publickey
}

// setter for public key
func (accnt *BaseAccount) SetPublicKey(pubKey crypto.PubKey) error {
	accnt.Publickey = pubKey
	return nil
}

// getter for nonce
func (accnt *BaseAccount) GetNonce() int64 {
	return accnt.Nonce
}

// setter for nonce
func (accnt *BaseAccount) SetNonce(nonce int64) error {
	accnt.Nonce = nonce
	return nil
}
