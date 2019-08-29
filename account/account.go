package account

import (
	"encoding/json"

	"github.com/QOSGroup/qbase/types"
	"github.com/tendermint/tendermint/crypto"
)

type Account interface {
	GetAddress() types.AccAddress
	SetAddress(addr types.AccAddress) error
	GetPublicKey() crypto.PubKey
	SetPublicKey(pubKey crypto.PubKey) error
	GetNonce() int64
	SetNonce(nonce int64) error
}

type BaseAccount struct {
	AccountAddress types.AccAddress `json:"account_address"` // account address
	Publickey      crypto.PubKey    `json:"public_key"`      // public key
	Nonce          int64            `json:"nonce"`           // identifies tx_status of an account
}

type jsonifyBaseAccount struct {
	AccountAddress string `json:"account_address"`
	Publickey      string `json:"public_key"`
	Nonce          int64  `json:"nonce"`
}

func ProtoBaseAccount() Account {
	return &BaseAccount{}
}

// getter for account address
func (accnt *BaseAccount) GetAddress() types.AccAddress {
	return accnt.AccountAddress
}

// setter for account address
func (accnt *BaseAccount) SetAddress(addr types.AccAddress) error {
	if addr.Empty() {
		return types.ErrInvalidAddress(types.CodeToDefaultMsg(types.CodeInvalidAddress))
	}
	accnt.AccountAddress = addr
	return nil
}

// getter for public key
func (accnt *BaseAccount) GetPublicKey() crypto.PubKey {
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

func (accnt *BaseAccount) MarshalJSON() ([]byte, error) {

	var pk string
	if accnt.Publickey != nil {
		pk = types.MustAccPubKeyString(accnt.Publickey)
	}

	acc := jsonifyBaseAccount{
		AccountAddress: accnt.AccountAddress.String(),
		Publickey:      pk,
		Nonce:          accnt.Nonce,
	}

	return json.Marshal(acc)
}

func (accnt *BaseAccount) UnmarshalJSON(data []byte) error {
	var acc jsonifyBaseAccount
	err := json.Unmarshal(data, &acc)
	if err != nil {
		return err
	}

	addr, err := types.AccAddressFromBech32(acc.AccountAddress)
	if err != nil {
		return err
	}

	var pk crypto.PubKey
	if len(acc.Publickey) != 0 {
		pk, err = types.GetAccPubKeyBech32(acc.Publickey)
		if err != nil {
			return err
		}
	}

	accnt.AccountAddress = addr
	accnt.Publickey = pk
	accnt.Nonce = acc.Nonce

	return nil
}
