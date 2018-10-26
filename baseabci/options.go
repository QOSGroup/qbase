package baseabci

import (
	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/mapper"
	"github.com/QOSGroup/qbase/qcp"
	"github.com/QOSGroup/qbase/store"
	"github.com/tendermint/tendermint/crypto"
)

func (app *BaseApp) SetName(name string) {
	if app.sealed {
		panic("SetName() on sealed BaseApp")
	}
	app.name = name
}

func (app *BaseApp) SetInitChainer(initChainer InitChainHandler) {
	if app.sealed {
		panic("SetInitChainer() on sealed BaseApp")
	}
	app.initChainer = initChainer
}

func (app *BaseApp) SetBeginBlocker(beginBlocker BeginBlockHandler) {
	if app.sealed {
		panic("SetBeginBlocker() on sealed BaseApp")
	}
	app.beginBlocker = beginBlocker
}

func (app *BaseApp) SetEndBlocker(endBlocker EndBlockHandler) {
	if app.sealed {
		panic("SetEndBlocker() on sealed BaseApp")
	}
	app.endBlocker = endBlocker
}

//RegisterQcpMapper 注册qcpMapper,
func (app *BaseApp) registerQcpMapper() {
	if app.sealed {
		panic("RegisterQcpMapper() on sealed BaseApp")
	}
	mapper := qcp.NewQcpMapper(app.GetCdc())
	app.RegisterMapper(mapper)
}

//RegisterQcpMapper 注册AccountMapper
func (app *BaseApp) RegisterAccountProto(proto func() account.Account) {
	if app.sealed {
		panic("RegisterAccountProto() on sealed BaseApp")
	}
	mapper := account.NewAccountMapper(app.GetCdc(), proto)
	app.RegisterMapper(mapper)
}

func (app *BaseApp) RegisterMapper(seedMapper mapper.IMapper) {
	if app.sealed {
		panic("RegisterMapper() on sealed BaseApp")
	}

	key := seedMapper.GetStoreKey()
	kvKey := key.(*store.KVStoreKey)
	app.mountStoresIAVL(kvKey)

	if _, ok := app.registerMappers[seedMapper.GetKVStoreName()]; ok {
		panic("Register dup mapper")
	}

	seedMapper.SetCodec(app.GetCdc())
	app.registerMappers[seedMapper.GetKVStoreName()] = seedMapper
}

func (app *BaseApp) RegisterCustomQueryHandler(handler CustomQueryHandler) {
	if app.sealed {
		panic("RegisterCustomQueryHandler() on sealed BaseApp")
	}
	app.customQueryHandler = handler
}

func (app *BaseApp) Seal()          { app.sealed = true }
func (app *BaseApp) IsSealed() bool { return app.sealed }
func (app *BaseApp) enforceSeal() {
	if !app.sealed {
		panic("enforceSeal() on BaseApp but not sealed")
	}
}

//-------------------------------------------------------------------

func (app *BaseApp) RegisterTxQcpSigner(signer crypto.PrivKey) {
	if app.sealed {
		panic("RegisterTxQcpSigner() on sealed BaseApp")
	}
	app.txQcpSigner = signer
}

func (app *BaseApp) RegisterTxQcpResultHandler(txQcpResultHandler TxQcpResultHandler) {
	if app.sealed {
		panic("RegisterTxQcpResultHandler() on sealed BaseApp")
	}
	app.txQcpResultHandler = txQcpResultHandler
}
