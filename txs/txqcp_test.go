package txs

import (
	"github.com/QOSGroup/qbase/types"
	"testing"
	"github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/stretchr/testify/require"
)

func TestTxQcp(t *testing.T) {
	var ext []common.KVPair
	ext = append(ext, common.KVPair{[]byte("k0"), []byte("v0")})
	ext = append(ext, common.KVPair{[]byte("k1"), []byte("v1")})

	txrst := NewQcpTxResult(0, ext, 10, types.NewInt(10), "test info")
	txstd := NewTxStd(txrst, "", types.NewInt(100))
	sig := Signature{
		ed25519.GenPrivKey().PubKey(),
		txstd.GetSignData(),
		1,
	}
	txqcp := NewTxQCP(txstd, "qsc1", "qos", 1, sig, 1, 2, false)
	require.NotNil(t, txqcp)
	txqcp.ValidateBasicData(true,"qsc1")
	data := txqcp.GetSigData()
	require.NotNil(t, data)
}