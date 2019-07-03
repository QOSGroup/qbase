package qcp

import (
	"encoding/hex"

	"github.com/QOSGroup/qbase/txs"
	"github.com/tendermint/tendermint/crypto"
)

const (
	EventModule = "qcp"
	From        = "qcp-from"
	To          = "qcp-to"
	Sequence    = "qcp-sequence"
	Hash        = "qcp-hash"
)

func GenQcpTxHash(tx *txs.TxQcp) string {
	bz := crypto.Sha256(tx.BuildSignatureBytes())
	return hex.EncodeToString(bz)
}
