package account

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/QOSGroup/qbase/types"
	"github.com/stretchr/testify/require"
	go_amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"
)

func keyPubAddr() (crypto.PrivKey, crypto.PubKey, types.AccAddress) {
	key := ed25519.GenPrivKey()
	pub := key.PubKey()
	addr := types.AccAddress(pub.Address())
	return key, pub, addr
}

func MakeCdc() *go_amino.Codec {
	cdc := go_amino.NewCodec()
	cryptoAmino.RegisterAmino(cdc)
	RegisterCodec(cdc)

	return cdc
}

func TestAccountMarshal(t *testing.T) {

	cdc := MakeCdc()
	types.RegisterCodec(cdc)

	_, pub, addr := keyPubAddr()
	baseAccount := BaseAccount{addr, nil, 0}

	err := baseAccount.SetPublicKey(pub)
	require.Nil(t, err)
	err = baseAccount.SetNonce(int64(7))
	require.Nil(t, err)

	add_binary, err := cdc.MarshalBinaryLengthPrefixed(baseAccount)
	require.Nil(t, err)

	another_add := BaseAccount{}
	another_json := []byte{}
	err = cdc.UnmarshalBinaryLengthPrefixed(add_binary, &another_add)
	require.Nil(t, err)
	require.Equal(t, baseAccount, another_add)
	// error on bad bytes
	another_add = BaseAccount{}
	another_json = []byte{}
	err = cdc.UnmarshalBinaryLengthPrefixed(add_binary[:len(add_binary)/2], &another_json)
	require.NotNil(t, err)

	//test json marshal
	var a BaseAccount
	data, e := json.Marshal(a)
	require.Nil(t, e)

	e = json.Unmarshal(data, &a)
	require.Nil(t, e)

	a1 := &BaseAccount{
		AccountAddress: addr,
	}

	data, e = json.Marshal(a1)
	require.Nil(t, e)

	e = json.Unmarshal(data, &a)
	require.Nil(t, e)
	require.Equal(t, a1.GetAddress(), a.GetAddress())

	a1 = &BaseAccount{
		AccountAddress: addr,
		Publickey:      pub,
	}

	data, e = json.Marshal(a1)
	require.Nil(t, e)

	e = json.Unmarshal(data, &a)
	require.Nil(t, e)
	require.Equal(t, a1.GetAddress(), a.GetAddress())
	require.Equal(t, a1.GetPublicKey(), a.GetPublicKey())

	a1 = &BaseAccount{
		AccountAddress: addr,
		Publickey:      pub,
		Nonce:          int64(1001),
	}

	data, e = json.Marshal(a1)
	require.Nil(t, e)

	e = json.Unmarshal(data, &a)
	require.Nil(t, e)
	require.Equal(t, a1.GetAddress(), a.GetAddress())
	require.Equal(t, a1.GetPublicKey(), a.GetPublicKey())
	require.Equal(t, a1.GetNonce(), a.GetNonce())

}

func TestBaseAccount_GetAddress(t *testing.T) {

	type appAccount struct {
		BaseAccount
		Amount int64
	}

	aa := appAccount{
		BaseAccount: BaseAccount{
			AccountAddress: types.AccAddress{},
			Publickey:      nil,
			Nonce:          10,
		},
		Amount: 20,
	}

	aa.SetNonce(30)

	fmt.Println(aa)
	fmt.Println(aa.BaseAccount)
}
