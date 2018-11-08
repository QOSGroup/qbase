package client

import (
	"encoding/hex"
	bacc "github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/example/basecoin/tx"
	"github.com/QOSGroup/qbase/txs"
	btx "github.com/QOSGroup/qbase/txs"
	btypes "github.com/QOSGroup/qbase/types"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

const (
	DeafultChainId = "basecoin"
)

func BuildStdTx(cdc *amino.Codec, from bacc.Account, hexPriKey string, to btypes.Address, coin btypes.BaseCoin) *btx.TxStd {
	sendTx := tx.NewSendTx(from.GetAddress(), to, coin)
	tx := txs.NewTxStd(&sendTx, DeafultChainId, btypes.NewInt(int64(0)))
	priHex, _ := hex.DecodeString(hexPriKey[2:])
	var priKey ed25519.PrivKeyEd25519
	cdc.MustUnmarshalBinaryBare(priHex, &priKey)
	signature, _ := tx.SignTx(priKey, from.GetNonce()+1)
	tx.Signature = []txs.Signature{txs.Signature{
		Pubkey:    priKey.PubKey(),
		Signature: signature,
		Nonce:     from.GetNonce() + 1,
	}}

	return tx
}

func BuildQCPTx(cdc *amino.Codec, std *btx.TxStd, chainId string, hexPriKey string, seq int64) *btx.TxQcp {
	tx := txs.NewTxQCP(std, chainId, DeafultChainId, seq, 0, 0, false, "")
	caHex, _ := hex.DecodeString(hexPriKey[2:])
	var caPriKey ed25519.PrivKeyEd25519
	cdc.MustUnmarshalBinaryBare(caHex, &caPriKey)
	sig, _ := tx.SignTx(caPriKey)
	tx.Sig.Nonce = seq
	tx.Sig.Signature = sig
	tx.Sig.Pubkey = caPriKey.PubKey()
	return tx
}
