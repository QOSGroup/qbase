package app

import (
	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/baseabci"
	"github.com/QOSGroup/qbase/context"
	ctx "github.com/QOSGroup/qbase/context"
	"github.com/QOSGroup/qbase/example/basecoin/tx"
	"github.com/QOSGroup/qbase/example/basecoin/types"
	"github.com/QOSGroup/qbase/qcp"
	qbtypes "github.com/QOSGroup/qbase/types"
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

	// txResult handler
	app.RegisterTxQcpResultHandler(resultTxHandler)

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

	accountMapper := ctx.Mapper(account.AccountMapperName).(*account.AccountMapper)
	qcpMapper := ctx.Mapper(qcp.QcpMapperName).(*qcp.QcpMapper)

	stateJSON := req.AppStateBytes
	genesisState := &types.GenesisState{}
	err := accountMapper.GetCodec().UnmarshalJSON(stateJSON, genesisState)
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

func resultTxHandler(ctx ctx.Context, txQcpResult interface{}) qbtypes.Result {
	// TODO
	return qbtypes.Result{}
}
