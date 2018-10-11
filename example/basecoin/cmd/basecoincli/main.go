package main

import (
	"flag"
	"fmt"
	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/example/basecoin/app"
	"github.com/QOSGroup/qbase/example/basecoin/tx"
	bctypes "github.com/QOSGroup/qbase/example/basecoin/types"
	"github.com/QOSGroup/qbase/txs"
	"github.com/QOSGroup/qbase/types"
	"github.com/tendermint/tendermint/rpc/client"
	"strconv"
	"strings"
	"time"
)

// send: go run main.go -m transfer -from xxx -to xxx -coin xxx,xxx
// query: go run main.go -m query -addr xxx
//
func main() {
	cdc := app.MakeCodec()

	mode := flag.String("m", "", "client mode: get/send")
	addr := flag.String("addr", "", "input account addr(bech32)")
	sender := flag.String("from", "", "input sender addr")
	receiver := flag.String("to", "", "input receive addr")
	coinStr := flag.String("coin", "", "input coinname,coinamount")

	flag.Parse()

	http := client.NewHTTP("tcp://127.0.0.1:26657", "/websocket")

	if *mode == "query" {
		if *addr == "" {
			panic("usage: go run main.go -m query -addr xxx ")
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

	if *mode == "transfer" {
		coin := strings.Split(*coinStr, ",")
		if *sender == "" || *receiver == "" || len(coin) != 2 {
			panic("usage: go run main.go -m transfer  -from xxx -to xxx -coin xxx,xxx")
		}
		senderAddr, _ := types.GetAddrFromBech32(*sender)
		receiverAddr, _ := types.GetAddrFromBech32(*receiver)
		amount, _ := strconv.ParseInt(coin[1], 10, 64)
		txStd := genSendTx(senderAddr, receiverAddr, bctypes.Coin{
			coin[0],
			types.NewInt(amount),
		})

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

}

func genSendTx(sender types.Address, receiver types.Address, coin bctypes.Coin) *txs.TxStd {
	sendTx := tx.SendTx{
		From:      sender,
		To:        receiver,
		Coin:      coin,
		Timestamp: time.Now().UnixNano(),
	}
	return txs.NewTxStd(&sendTx, "basecoin-chain", types.NewInt(int64(0)))
}
