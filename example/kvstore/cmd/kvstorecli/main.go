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

// 查询Key值:
//   1.  go run main.go -m get -k xxx
//   2.  go run main.go -k xxx

// 设置Key-Value键值对:
//   1.  go run main.go -m set -k xxx -v xxx
//   2.  go run main.go -m send -k xxx -v xxx

func main() {

	cdc := kvstore.MakeKVStoreCodec()

	mode := flag.String("m", "", "client mode \n: get , set  or  send")
	key := flag.String("k", "", "input key")
	value := flag.String("v", "", "input value")

	flag.Parse()

	http := client.NewHTTP("tcp://127.0.0.1:26657", "/websocket")

	g, _ := http.Genesis()
	chainID := g.Genesis.ChainID
	fmt.Println(fmt.Sprintf("current chainID is %s.", chainID))

	switch *mode {
	case "get", "":
		if *key == "" {
			panic("usage: go run main.go -m get -k xxx ")
		}
		v := getValue(*key, http, cdc)
		fmt.Println(fmt.Sprintf("query kv result: %s=%s", *key, v))
	case "send", "set":
		if *key == "" || *value == "" {
			panic("usage: go run main.go -m set  -k xxx -v xxx")
		}

		sendKVTx(*key, *value, chainID, http, cdc)
		fmt.Println(fmt.Sprintf("set kv: %s = %s", *key, *value))
	default:
		panic("wrong mode")
	}

}

func getValue(key string, http *client.HTTP, cdc *go_amino.Codec) string {
	result, err := http.ABCIQuery("/store/kv/key", []byte(key))
	if err != nil {
		panic(err)
	}

	queryValueBz := result.Response.GetValue()
	if queryValueBz == nil {
		return ""
	}
	var queryValue string
	cdc.UnmarshalBinaryBare(queryValueBz, &queryValue)

	return queryValue
}

func sendKVTx(k, v, chainID string, http *client.HTTP, cdc *go_amino.Codec) {

	txStd := wrapToStdTx(k, v, chainID)

	tx, err := cdc.MarshalBinaryBare(txStd)
	if err != nil {
		panic(err)
	}

	_, err = http.BroadcastTxSync(tx)
	if err != nil {
		panic(err)
	}
}

func wrapToStdTx(key, value, chainID string) *txs.TxStd {
	kv := kvstore.NewKvstoreTx([]byte(key), []byte(value))
	return txs.NewTxStd(kv, chainID, types.NewUint(uint64(10000)))
}
