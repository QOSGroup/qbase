package main

import (
	"github.com/QOSGroup/qbase/example/inittest"
	"github.com/QOSGroup/qbase/txs"
	"github.com/QOSGroup/qbase/types"
	go_amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/rpc/client"
)

func main() {

	key := "a"
	val := "Long live imuge!"

	cdc := inittest.MakeCodec()

	bz := getBytes(key, val, cdc)

	http := client.NewHTTP("tcp://127.0.0.1:26657", "/websocket")

	http.BroadcastTxAsync(bz)

}

func getBytes(key string, value string, cdc *go_amino.Codec) []byte {
	kv := inittest.NewInitTestTx([]byte(key), []byte(value))
	txStd := txs.NewTxStd(kv, "kv-chain", types.NewInt(int64(10000)))

	bz := cdc.MustMarshalBinaryBare(txStd)

	return bz
}
