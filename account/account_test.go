package account

import (
	"testing"

	"github.com/QOSGroup/qbase/types"
	"github.com/stretchr/testify/require"
	go_amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/encoding/amino"
)

func keyPubAddr() (crypto.PrivKey, crypto.PubKey, types.Address) {
	key := ed25519.GenPrivKey()
	pub := key.PubKey()
	addr := types.Address(pub.Address())
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

	_, pub, addr := keyPubAddr()
	baseAccount := BaseAccount{addr, nil, 0}

	err := baseAccount.SetPublicKey(pub)
	require.Nil(t, err)
	err = baseAccount.SetNonce(int64(7))
	require.Nil(t, err)

	add_binary, err := cdc.MarshalBinary(baseAccount)
	require.Nil(t, err)

	another_add := BaseAccount{}
	another_json := []byte{}
	err = cdc.UnmarshalBinary(add_binary, &another_add)
	require.Nil(t, err)
	require.Equal(t, baseAccount, another_add)

	// error on bad bytes
	another_add = BaseAccount{}
	another_json = []byte{}
	err = cdc.UnmarshalBinary(add_binary[:len(add_binary)/2], &another_json)
	require.NotNil(t, err)

}
