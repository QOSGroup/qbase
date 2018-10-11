package main

import (
	"flag"
	"fmt"

	"github.com/QOSGroup/qbase/example/kvstore"
	"github.com/QOSGroup/qbase/txs"
	"github.com/QOSGroup/qbase/types"
	go_amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/rpc/client"
)

// send: go run main.go -m send -k xxx -v xxx
// query: go run main.go -m get -k xxx
//
func main() {

	cdc := MakeCdc()

	mode := flag.String("m", "", "client mode: get/send")
	key := flag.String("k", "", "input key")
	value := flag.String("v", "", "input value")

	flag.Parse()

	http := client.NewHTTP("tcp://127.0.0.1:26657", "/websocket")

	if *mode == "get" {
		if *key == "" {
			panic("usage: go run main.go -m get -key xxx ")
		}

		result, err := http.ABCIQuery("/store/kv/key", []byte(*key))
		if err != nil {
			panic(err)
		}

		queryValueBz := result.Response.GetValue()
		var queryValue string
		cdc.UnmarshalBinaryBare(queryValueBz, &queryValue)

		fmt.Println(fmt.Sprintf("query kv is %s = %s", *key, queryValue))
	}

	if *mode == "send" {
		if *key == "" || *value == "" {
			panic("usage: go run main.go -m send  -key xxx -value xxx")
		}

		txStd := wrapToStdTx(*key, *value)

		tx, err := cdc.MarshalBinaryBare(txStd)
		if err != nil {
			panic("use cdc encode object fail")
		}

		_, err = http.BroadcastTxSync(tx)
		if err != nil {
			fmt.Println(err)
			panic("BroadcastTxSync err")
		}

		fmt.Println(fmt.Sprintf("send kv is %s = %s", *key, *value))
	}

}

func wrapToStdTx(key string, value string) *txs.TxStd {
	kv := kvstore.NewKvstoreTx([]byte(key), []byte(value))
	return txs.NewTxStd(kv, "kv-chain", types.NewInt(int64(10000)))
}

func MakeCdc() *go_amino.Codec {
	var cdc = go_amino.NewCodec()
	txs.RegisterCodec(cdc)

	cdc.RegisterConcrete(&kvstore.KvstoreTx{}, "kvstore/main/kvstoretx", nil)
	return cdc
}
