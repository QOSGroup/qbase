package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/example/basecoin/app"
	"github.com/QOSGroup/qbase/example/basecoin/tx"
	bctypes "github.com/QOSGroup/qbase/example/basecoin/types"
	"github.com/QOSGroup/qbase/txs"
	"github.com/QOSGroup/qbase/types"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/rpc/client"
	"strconv"
	"strings"
)

func main() {
	cdc := app.MakeCodec()

	mode := flag.String("m", "", "client mode: get/send")
	addr := flag.String("addr", "", "input account addr(bech32)")
	sender := flag.String("from", "", "input sender addr")
	receiver := flag.String("to", "", "input receive addr")
	prikey := flag.String("prikey", "", "input sender prikey")
	nonce := flag.Int64("nonce", 0, "input sender nonce")
	coinStr := flag.String("coin", "", "input coinname,coinamount")
	chainId := flag.String("chainid", "", "input qcp chainId")
	qcpPriKey := flag.String("qcpprikey", "", "input qcp prikey")
	qcpseq := flag.Int64("qcpseq", 0, "input qcp sequence")
	isresult := flag.Bool("isresult", false, "input qcp isresult")

	flag.Parse()

	http := client.NewHTTP("tcp://127.0.0.1:26657", "/websocket")

	switch *mode {
	case "accquery": // 账户查询
		queryAccount(http, cdc, addr)
	case "qcpseq": // QCP sequence查询
		queryQCPSequence(http, cdc, chainId, qcpseq)
	case "qcpquery": // QCP查询
		queryQCP(http, cdc, chainId, qcpseq)
	case "stdtransfer": // 链内交易
		stdTransfer(http, cdc, sender, prikey, receiver, coinStr, nonce)
	case "qcptransfer": // QCP交易
		qcpTransfer(http, cdc, sender, prikey, receiver, coinStr, nonce, chainId, qcpPriKey, qcpseq, isresult)
	}

}

// 查询账户状态
func queryAccount(http *client.HTTP, cdc *amino.Codec, addr *string) {
	if *addr == "" {
		panic("usage: -m=accquery -addr=xxx")
	}
	address, _ := types.GetAddrFromBech32(*addr)
	key := account.AddressStoreKey(address)
	result, err := http.ABCIQuery("/store/acc/key", key)
	if err != nil {
		panic(err)
	}

	queryValueBz := result.Response.GetValue()
	var acc *bctypes.AppAccount
	cdc.UnmarshalBinaryBare(queryValueBz, &acc)

	fmt.Println(fmt.Sprintf("query addr is %s = %v", *addr, acc))
}

// 查询QCP Sequence
func queryQCPSequence(http *client.HTTP, cdc *amino.Codec, chainid *string, qcpseq *int64) {
	if *chainid == "" {
		panic("usage: -m=qcpseq -chainid=xxx")
	}

	// in sequence
	keyIn := fmt.Sprintf("[%s]/in/sequence", *chainid)
	inResult, err := http.ABCIQuery("/store/qcp/key", []byte(keyIn))
	if err != nil {
		panic(err)
	}
	var in int64
	if inResult.Response.GetValue() != nil {
		cdc.UnmarshalBinaryBare(inResult.Response.GetValue(), &in)
	}

	// out sequence
	keyOut := fmt.Sprintf("[%s]/out/sequence", *chainid)
	outResult, err := http.ABCIQuery("/store/qcp/key", []byte(keyOut))
	if err != nil {
		panic(err)
	}
	var out int64
	if outResult.Response.GetValue() != nil {
		cdc.UnmarshalBinaryBare(outResult.Response.GetValue(), &out)
	}

	fmt.Println(fmt.Sprintf("query chain is %s, sequence in/out: %d/%d", *chainid, in, out))
}

// 查询QCP状态
func queryQCP(http *client.HTTP, cdc *amino.Codec, chainid *string, qcpseq *int64) {
	if *chainid == "" || *qcpseq <= 0 {
		panic("usage: -m=qcpquery -chainid=xxx -qcpseq=xxx -inout=xxx")
	}

	key := fmt.Sprintf("[%s]/out/tx_[%d]", *chainid, *qcpseq)
	result, err := http.ABCIQuery("/store/qcp/key", []byte(key))
	if err != nil {
		panic(err)
	}

	var tx txs.TxQcp
	if result.Response.GetValue() != nil {
		cdc.UnmarshalBinaryBare(result.Response.GetValue(), &tx)
	}

	fmt.Println(fmt.Sprintf("query chain is %s, tx out[%d] is %v", *chainid, *qcpseq, tx))
}

// 链内交易
func stdTransfer(http *client.HTTP, cdc *amino.Codec, sender *string, prikey *string, receiver *string, coinStr *string, nonce *int64) {
	coin := strings.Split(*coinStr, ",")
	if *sender == "" || *receiver == "" || len(coin) != 2 || *prikey == "" || *nonce < 0 {
		panic("usage: -m=stdTransfer -from=xxx -to=xxx -coin=xxx,xxx -prikey=xxx -nonce=xxx(>=0)")
	}
	senderAddr, _ := types.GetAddrFromBech32(*sender)
	receiverAddr, _ := types.GetAddrFromBech32(*receiver)
	amount, _ := strconv.ParseInt(coin[1], 10, 64)
	txStd := genStdSendTx(cdc, senderAddr, receiverAddr, bctypes.Coin{
		coin[0],
		types.NewInt(amount),
	}, *prikey, *nonce)

	tx, err := cdc.MarshalBinaryBare(txStd)
	if err != nil {
		panic("use cdc encode object fail")
	}

	_, err = http.BroadcastTxSync(tx)
	if err != nil {
		fmt.Println(err)
		panic("BroadcastTxSync err")
	}

	fmt.Println(fmt.Sprintf("send tx is %v", txStd))
}

// QCP交易
func qcpTransfer(http *client.HTTP, cdc *amino.Codec, sender *string, prikey *string, receiver *string, coinStr *string, nonce *int64,
	chainId *string, qcpPriKey *string, qcpseq *int64, isresult *bool) {
	coin := strings.Split(*coinStr, ",")
	if *sender == "" || *receiver == "" || len(coin) != 2 || *nonce < 0 || *chainId == "" || *qcpPriKey == "" || *qcpseq <= 0 {
		panic("usage: -m=qcpTransfer -from=xxx -to=xxx -coin=xxx,xxx -prikey=xxx -nonce=xxx(>=0) -chainid=xxx -qcpprikey=xxx -qcpseq=xxx")
	}
	senderAddr, _ := types.GetAddrFromBech32(*sender)
	receiverAddr, _ := types.GetAddrFromBech32(*receiver)
	amount, _ := strconv.ParseInt(coin[1], 10, 64)
	txStd := genQcpSendTx(cdc, senderAddr, receiverAddr, bctypes.Coin{
		coin[0],
		types.NewInt(amount),
	}, *prikey, *nonce, *chainId, *qcpPriKey, *qcpseq, *isresult)

	tx, err := cdc.MarshalBinaryBare(txStd)
	if err != nil {
		panic("use cdc encode object fail")
	}

	_, err = http.BroadcastTxSync(tx)
	if err != nil {
		fmt.Println(err)
		panic("BroadcastTxSync err")
	}

	fmt.Println(fmt.Sprintf("send tx is %v", txStd))
}

// 生成链内交易SendTx
func genStdSendTx(cdc *amino.Codec, sender types.Address, receiver types.Address, coin bctypes.Coin,
	senderPriHex string, nonce int64) *txs.TxStd {
	sendTx := tx.NewSendTx(sender, receiver, coin)
	tx := txs.NewTxStd(&sendTx, "basecoin-chain", types.NewInt(int64(0)))
	priHex, _ := hex.DecodeString(senderPriHex[2:])
	var priKey ed25519.PrivKeyEd25519
	cdc.MustUnmarshalBinaryBare(priHex, &priKey)
	signData := append(tx.GetSignData(), types.Int2Byte(int64(nonce))...)
	signature, _ := priKey.Sign(signData)
	tx.Signature = []txs.Signature{txs.Signature{
		Pubkey:    priKey.PubKey(),
		Signature: signature,
		Nonce:     nonce,
	}}

	return tx
}

// 生成QCP交易SendTx
func genQcpSendTx(cdc *amino.Codec, sender types.Address, receiver types.Address, coin bctypes.Coin,
	senderPriHex string, nonce int64, chainId string, caPriHex string, qcpseq int64, isresult bool) *txs.TxQcp {
	sendTx := tx.NewSendTx(sender, receiver, coin)
	std := txs.NewTxStd(&sendTx, "basecoin-chain", types.NewInt(int64(0)))
	priHex, _ := hex.DecodeString(senderPriHex[2:])
	var priKey ed25519.PrivKeyEd25519
	cdc.MustUnmarshalBinaryBare(priHex, &priKey)
	signData := append(std.GetSignData(), types.Int2Byte(int64(nonce))...)
	signature, _ := priKey.Sign(signData)
	std.Signature = []txs.Signature{txs.Signature{
		Pubkey:    priKey.PubKey(),
		Signature: signature,
		Nonce:     nonce,
	}}
	tx := txs.NewTxQCP(std, chainId, "basecoin-chain", qcpseq, 0, 0, isresult)
	caHex, _ := hex.DecodeString(caPriHex[2:])
	var caPriKey ed25519.PrivKeyEd25519
	cdc.MustUnmarshalBinaryBare(caHex, &caPriKey)
	sig, _ := caPriKey.Sign(tx.GetSigData())
	tx.Sig.Nonce = qcpseq
	tx.Sig.Signature = sig
	tx.Sig.Pubkey = caPriKey.PubKey()
	return tx
}
