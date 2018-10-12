package app

import (
	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/baseabci"
	"github.com/QOSGroup/qbase/context"
	"github.com/QOSGroup/qbase/example/basecoin/tx"
	"github.com/QOSGroup/qbase/example/basecoin/types"
	"github.com/tendermint/go-amino"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/encoding/amino"
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

	baseApp := baseabci.NewBaseApp(appName, logger, db, registerCdc)
	baseApp.SetCommitMultiStoreTracer(traceStore)

	app := &BaseCoinApp{
		BaseApp: baseApp,
	}

	//设置 InitChainer
	app.SetInitChainer(app.initChainer)

	//账户mapper
	app.RegisterAccountProto(types.NewAppAccount)

	// Mount stores and load the latest state.
	err := app.LoadLatestVersion()
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}

// 初始配置
func (app *BaseCoinApp) initChainer(ctx context.Context, req abci.RequestInitChain) abci.ResponseInitChain {

	accountMapper := ctx.Mapper(account.AccountMapperName).(*account.AccountMapper)

	stateJSON := req.AppStateBytes
	genesisState := &types.GenesisState{}
	err := accountMapper.GetCodec().UnmarshalJSON(stateJSON, genesisState)
	if err != nil {
		panic(err)
	}

	//保存初始账户
	for _, gacc := range genesisState.Accounts {
		acc, err := gacc.ToAppAccount()
		if err != nil {
			panic(err)
		}
		accountMapper.SetAccount(acc)
	}

	return abci.ResponseInitChain{}
}

// 序列化反序列化相关注册
func MakeCodec() *amino.Codec {
	var cdc = amino.NewCodec()
	cryptoAmino.RegisterAmino(cdc)
	registerCdc(cdc)
	baseabci.RegisterCodec(cdc)
	return cdc
}

func registerCdc(cdc *amino.Codec) {
	cdc.RegisterConcrete(&types.AppAccount{}, "basecoin/AppAccount", nil)
	cdc.RegisterConcrete(&tx.SendTx{}, "basecoin/SendTx", nil)
}
