package txs

import (
	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/common"
	"testing"
)

func TestTxQcp(t *testing.T) {
	var ext []common.KVPair
	ext = append(ext, common.KVPair{[]byte("k0"), []byte("v0")})
	ext = append(ext, common.KVPair{[]byte("k1"), []byte("v1")})

	txrst := NewQcpTxResult(0, ext, 10, types.NewInt(10), "qcp result info")
	txstd := NewTxStd(txrst, "qsc1", types.NewInt(100))

	signer := txstd.ITx.GetSigner()
	accmapper := account.NewAccountMapper(account.ProtoBaseAccount)

	//填充 txstd.Signature[]
	for _, sg := range signer {
		prvKey := ed25519.GenPrivKey()
		nonce, err := accmapper.GetNonce(sg)
		if err != nil {
			t.Errorf("GetNonce() for address(%s) error", string(sg))
		}
		txstd.SignTx(prvKey, int64(nonce))
	}

	txqcp := NewTxQCP(txstd, "qsc1", "qos", 1, 13452345, 2, false)
	txqcp.SignTx(ed25519.GenPrivKey(), 1)
	require.NotNil(t, txqcp)
	txqcp.ValidateBasicData(true, "qsc1")
	data := txqcp.GetSigData()
	require.NotNil(t, data)
}
