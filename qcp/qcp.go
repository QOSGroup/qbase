package qcp

import (
	"encoding/hex"

	"github.com/QOSGroup/qbase/txs"
	"github.com/tendermint/tendermint/crypto"
)

const (
	QcpFrom     = "qcp.from"
	QcpTo       = "qcp.to"
	QcpSequence = "qcp.sequence"
	QcpHash     = "qcp.hash"
)

func GenQcpTxHash(tx *txs.TxQcp) string {
	bz := crypto.Sha256(tx.BuildSignatureBytes())
	return hex.EncodeToString(bz)
}
