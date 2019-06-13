package app

import (
	"fmt"
	"io"

	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/store"
	btypes "github.com/QOSGroup/qbase/types"

	"github.com/QOSGroup/qbase/baseabci"
	"github.com/QOSGroup/qbase/context"
	"github.com/QOSGroup/qbase/example/basecoin/types"
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

const (
	appName = "basecoin"

	// how much gas = 1 qstar
	gasPerUnitCost = 1000
)

type BaseCoinApp struct {
	*baseabci.BaseApp
}

func NewApp(logger log.Logger, db dbm.DB, traceStore io.Writer) *BaseCoinApp {

	baseApp := baseabci.NewBaseApp(appName, logger, db, RegisterCodec, baseabci.SetPruning(store.PruneSyncable))
	baseApp.SetCommitMultiStoreTracer(traceStore)

	//baseApp.

	app := &BaseCoinApp{
		BaseApp: baseApp,
	}

	// 设置 InitChainer
	app.SetInitChainer(app.initChainer)

	app.SetGasHandler(app.gasHandler)

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

	accountMapper := baseabci.GetAccountMapper(ctx)

	stateJSON := req.AppStateBytes
	genesisState := &types.GenesisState{}
	err := app.BaseApp.GetCdc().UnmarshalJSON(stateJSON, genesisState)
	if err != nil {
		panic(err)
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

func (app *BaseCoinApp) gasHandler(ctx context.Context, payer btypes.Address) (gasUsed uint64, err btypes.Error) {
	gasFeeUsed := int64(ctx.GasMeter().GasConsumed()) / gasPerUnitCost

	if gasFeeUsed > 0 {
		accountMapper := ctx.Mapper(account.AccountMapperName).(*account.AccountMapper)
		account := accountMapper.GetAccount(payer).(*types.AppAccount)

		if account.Coins.AmountOf("qstar").Int64() < gasFeeUsed {
			log := fmt.Sprintf("%s no enough coins to pay the gas after this tx done", payer)
			return uint64(gasFeeUsed * gasPerUnitCost), btypes.ErrInternal(log)
		}

		account.Coins = account.Coins.Minus(btypes.BaseCoins{btypes.NewInt64BaseCoin("qstar", gasFeeUsed)})
		accountMapper.SetAccount(account)

	}

	return uint64(gasFeeUsed * gasPerUnitCost), nil
}
