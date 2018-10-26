package app

import (
	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/baseabci"
	"github.com/QOSGroup/qbase/context"
	"github.com/QOSGroup/qbase/example/basecoin/types"
	"github.com/QOSGroup/qbase/qcp"
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	"io"
)

const (
	appName = "basecoin"
)

type BaseCoinApp struct {
	*baseabci.BaseApp
}

func NewApp(logger log.Logger, db dbm.DB, traceStore io.Writer) *BaseCoinApp {

	baseApp := baseabci.NewBaseApp(appName, logger, db, RegisterCodec)
	baseApp.SetCommitMultiStoreTracer(traceStore)

	app := &BaseCoinApp{
		BaseApp: baseApp,
	}

	// 设置 InitChainer
	app.SetInitChainer(app.initChainer)

	// 账户mapper
	app.RegisterAccountProto(types.NewAppAccount)

	// QCP mapper
	// 默认已注入

	// Mount stores and load the latest state.
	err := app.LoadLatestVersion()
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}

// 初始配置
func (app *BaseCoinApp) initChainer(ctx context.Context, req abci.RequestInitChain) abci.ResponseInitChain {

	accountMapper := ctx.Mapper(account.GetAccountKVStoreName()).(*account.AccountMapper)
	qcpMapper := ctx.Mapper(qcp.GetQcpKVStoreName()).(*qcp.QcpMapper)

	stateJSON := req.AppStateBytes
	genesisState := &types.GenesisState{}
	err := app.BaseApp.GetCdc().UnmarshalJSON(stateJSON, genesisState)
	if err != nil {
		panic(err)
	}

	// 保存初始QCP配置
	for _, qcp := range genesisState.QCP {
		qcpMapper.SetChainInTruestPubKey(qcp.ChainId, qcp.PubKey)
	}

	// 保存初始账户
	for _, gacc := range genesisState.Accounts {
		acc, err := gacc.ToAppAccount()
		if err != nil {
			panic(err)
		}
		accountMapper.SetAccount(acc)
	}

	return abci.ResponseInitChain{}
}
