package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/baseabci"
	"github.com/QOSGroup/qbase/example/kvstore"
	"github.com/QOSGroup/qbase/store"

	go_amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/abci/server"
	"github.com/tendermint/tendermint/libs/log"

	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func main() {

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "main")
	rootDir := ""
	db, err := dbm.NewGoLevelDB("kvstore", filepath.Join(rootDir, "data"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var baseapp = baseabci.NewBaseApp("kvstore", logger, db, func(cdc *go_amino.Codec) {
		kvstore.RegisterCodec(cdc)
	})

	baseapp.RegisterAccountProto(func() account.Account {
		return &account.BaseAccount{}
	})

	var mainStore = store.NewKVStoreKey("kv")
	var kvMapper = kvstore.NewKvMapper(mainStore)
	baseapp.RegisterMapper(kvMapper)

	if err := baseapp.LoadLatestVersion(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Start the ABCI server
	srv, err := server.NewServer("0.0.0.0:26658", "socket", baseapp)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = srv.Start()
	if err != nil {
		cmn.Exit(err.Error())
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		err = srv.Stop()
		if err != nil {
			cmn.Exit(err.Error())
		}
	})
	return

}
