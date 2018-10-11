package inittest

import (
	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/baseabci"
	"github.com/QOSGroup/qbase/context"
	"github.com/QOSGroup/qbase/store"
	"github.com/tendermint/go-amino"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/encoding/amino"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	"io"
)

const (
	appName = "inittest"
)

type InitTestApp struct {
	*baseabci.BaseApp
}

func NewApp(logger log.Logger, db dbm.DB, traceStore io.Writer) *InitTestApp {

	baseApp := baseabci.NewBaseApp(appName, logger, db, registerCdc)
	baseApp.SetCommitMultiStoreTracer(traceStore)

	app := &InitTestApp{
		BaseApp: baseApp,
	}

	// Set InitChainer
	app.SetInitChainer(app.initChainer)

	//账户mapper
	app.RegisterAccount(func() account.Account {
		return &QOSAccount{}
	})

	baseStore := store.NewKVStoreKey("base")
	txMapper := NewTxMapper(baseStore)
	app.RegisterSeedMapper(txMapper)

	// Mount stores and load the latest state.
	err := app.LoadLatestVersion()
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}

// 初始配置
func (app *InitTestApp) initChainer(ctx context.Context, req abci.RequestInitChain) abci.ResponseInitChain {

	baseMapper := ctx.Mapper(TX_MAPPER_NAME)
	accountMapper := ctx.Mapper(account.AccountMapperName)

	//保存CA和初始账户
	stateJSON := req.AppStateBytes
	genesisState := &GenesisState{}
	err := app.GetCdc().UnmarshalJSON(stateJSON, genesisState)
	if err != nil {
		panic(err)
	}
	//保存CA
	baseMapper.SetObject([]byte("rootca"), genesisState.CAPubKey.Bytes())

	//保存账户
	for _, gacc := range genesisState.Accounts {
		acc, err := gacc.ToQosAccount()
		if err != nil {
			panic(err)
		}
		accountMapper.SetObject([]byte(acc.BaseAccount.AccountAddress), app.GetCdc().MustMarshalBinaryBare(acc))
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
	cdc.RegisterConcrete(&InitTestTx{}, "inittest/InitTestTx", nil)
	cdc.RegisterConcrete(&QOSAccount{}, "inittest/QOSAccount", nil)
}
