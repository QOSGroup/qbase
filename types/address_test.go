package types

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

var invalidStrs = []string{
	"",
	"hello, world!",
	"0xAA",
	"AAA",
	PREF_ADD + "AB0C",
	PREF_ADD + "1234",
}

func testMarshal(t *testing.T, original interface{}, res interface{}, marshal func() ([]byte, error), unmarshal func([]byte) error) {
	bz, err := marshal()
	require.Nil(t, err)
	err = unmarshal(bz)
	require.Nil(t, err)
	require.Equal(t, original, res)
}

func TestAddress(t *testing.T) {
	var pub ed25519.PubKeyEd25519

	for i := 0; i < 100; i++ {
		rand.Read(pub[:])

		acc := Address(pub.Address())
		res := Address{}

		testMarshal(t, &acc, &res, acc.MarshalJSON, (&res).UnmarshalJSON)
		testMarshal(t, &acc, &res, acc.Marshal, (&res).Unmarshal)

		str := acc.String()
		res, err := GetAddrFromBech32(str)
		require.Nil(t, err)
		require.Equal(t, acc, res)

		jsonStr, err := json.Marshal(str)
		require.Nil(t, err)

		res.UnmarshalJSON(jsonStr)
		require.Equal(t, acc, res)

	}

}