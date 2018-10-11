package server

import (
	"io"
	"os"
	"path/filepath"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

type (
	// AppCreator reflects a function that allows us to lazily initialize an
	// application using various configurations.
	AppCreator func(home string, logger log.Logger, traceStore string) (abci.Application, error)

	// AppCreatorInit reflects a function that performs initialization of an
	// AppCreator.
	AppCreatorInit func(log.Logger, dbm.DB, io.Writer) abci.Application
)

// ConstructAppCreator returns an application generation function.
func ConstructAppCreator(appFn AppCreatorInit, name string) AppCreator {
	return func(rootDir string, logger log.Logger, traceStore string) (abci.Application, error) {
		dataDir := filepath.Join(rootDir, "data")

		db, err := dbm.NewGoLevelDB(name, dataDir)
		if err != nil {
			return nil, err
		}

		var traceStoreWriter io.Writer
		if traceStore != "" {
			traceStoreWriter, err = os.OpenFile(
				traceStore,
				os.O_WRONLY|os.O_APPEND|os.O_CREATE,
				0666,
			)
			if err != nil {
				return nil, err
			}
		}

		app := appFn(logger, db, traceStoreWriter)
		return app, nil
	}
}
