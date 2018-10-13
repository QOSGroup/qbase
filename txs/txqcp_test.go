package txs

import (
	"fmt"
	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/common"
	"testing"
)

func newQcpTxResult() (ret *QcpTxResult) {
	var ext []common.KVPair
	ext = append(ext, common.KVPair{[]byte("k0"), []byte("v0")})
	ext = append(ext, common.KVPair{[]byte("k1"), []byte("v1")})

	ret = NewQcpTxResult(0, ext, 10, types.NewInt(10), "qcp result info")
	if !ret.ValidateData() {
		fmt.Print("QcpTxResult ValidateData Error")
		return nil
	}

	return
}

func newTxStd(tx ITx) (ret *TxStd) {
	txstd := NewTxStd(tx, "qsc1", types.NewInt(100))
	signer := txstd.ITx.GetSigner()
	err := txstd.ValidateBasicData(true)
	if err != nil {
		fmt.Print("TxStd ValidateData Error")
		return nil
	}

	accmapper := account.NewAccountMapper(account.ProtoBaseAccount)

	//填充 txstd.Signature[]
	for _, sg := range signer {
		prvKey := ed25519.GenPrivKey()
		nonce, err := accmapper.GetNonce(sg)
		if err != nil {
			fmt.Printf("GetNonce() for address(%s) error", string(sg))
			return nil
		}
		if !txstd.SignTx(prvKey, int64(nonce)) {
			fmt.Printf("SignTx(addr:%s) error", string(sg))
			return nil
		}
	}

	return
}

func TestNewQcpTxResult(t *testing.T) {
	txResult := newQcpTxResult()
	if txResult == nil {
		fmt.Print("New QcpTxResult error")
		return
	}

	if txResult.GetSigner() == nil {
		fmt.Print("No signer!")
	}

	if txResult.GetGasPayer() == nil {
		fmt.Print("No payer")
	}

	fmt.Printf("gas(%d)", txResult.CalcGas().Int64())
}

func TestNewTxStd(t *testing.T) {
	txResult := newQcpTxResult()
	if txResult == nil {
		t.Error("New QcpTx error!")
		return
	}

	txStd := newTxStd(txResult)
	if txResult == nil {
		t.Error("New TxStd error!")
		return
	}

	fmt.Printf("TxStd type: %s", txStd.Type())
}

//TxQcp test
func TestTxQcp(t *testing.T) {
	txResult := newQcpTxResult()
	if txResult == nil {
		t.Error("New QcpTx error!")
		return
	}
	txStd := newTxStd(txResult)
	if txResult == nil {
		t.Error("New TxStd error!")
		return
	}

	txqcp := NewTxQCP(txStd, "qsc1", "qos", 1, 13452345, 2, false)
	txqcp.SignTx(ed25519.GenPrivKey())
	require.NotNil(t, txqcp)
	err := txqcp.ValidateBasicData(true, "qsc1")
	if err != nil {
		t.Errorf("TxQCP ValidateData Error")
	}
	data := txqcp.GetSigData()
	require.NotNil(t, data)
}
