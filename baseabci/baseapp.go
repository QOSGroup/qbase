package baseabci

import (
	"bytes"
	"fmt"
	"io"
	"runtime/debug"
	"strings"

	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/qcp"

	"github.com/QOSGroup/qbase/mapper"

	ctx "github.com/QOSGroup/qbase/context"
	"github.com/QOSGroup/qbase/store"
	"github.com/QOSGroup/qbase/txs"
	"github.com/QOSGroup/qbase/types"

	go_amino "github.com/tendermint/go-amino"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/tmhash"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

// BaseApp reflects the ABCI application implementation.
type BaseApp struct {
	// initialized on creation
	Logger    log.Logger
	name      string                 // application name from abci.Info
	db        dbm.DB                 // common DB backend
	cms       store.CommitMultiStore // Main (uncached) state
	txDecoder types.TxDecoder        // unmarshal []byte into stdTx or qcpTx

	// may be nil
	initChainer  InitChainHandler  // initialize state with validators and state blob
	beginBlocker BeginBlockHandler // logic to run before any txs
	endBlocker   EndBlockHandler   // logic to run after all txs, and to determine valset changes

	//--------------------
	// Volatile
	// checkState is set on initialization and reset on Commit.
	// deliverState is set in InitChain and BeginBlock and cleared on Commit.
	checkState       *state                  // for CheckTx
	deliverState     *state                  // for DeliverTx
	signedValidators []abci.SigningValidator // absent validators from begin block

	//--------------------------------------------------------------

	//TODO: may be nil , 做校验
	txQcpResultHandler  TxQcpResultHandler // exec方法中回调，执行具体的业务逻辑
	signerForCrossTxQcp crypto.PrivKey     //对跨链TxQcp签名的私钥， app启动时初始化

	//注册的mapper
	registerMappers map[string]mapper.IMapper

	cdc *go_amino.Codec
	// flag for sealing
	sealed bool
}

var _ abci.Application = (*BaseApp)(nil)

// NewBaseApp returns a reference to an initialized BaseApp.
func NewBaseApp(name string, logger log.Logger, db dbm.DB, registerCodecFunc func(*go_amino.Codec), options ...func(*BaseApp)) *BaseApp {

	if registerCodecFunc != nil {
		registerCodecFunc(cdc)
	}

	txDecoder := types.GetTxDecoder(cdc)

	app := &BaseApp{
		Logger:          logger,
		name:            name,
		db:              db,
		cms:             store.NewCommitMultiStore(db),
		txDecoder:       txDecoder,
		cdc:             cdc,
		registerMappers: make(map[string]mapper.IMapper),
	}

	for _, option := range options {
		option(app)
	}

	app.registerQcpMapper()

	return app
}

// BaseApp Name
func (app *BaseApp) Name() string {
	return app.name
}

func (app *BaseApp) GetCdc() *go_amino.Codec {
	return app.cdc
}

// SetCommitMultiStoreTracer sets the store tracer on the BaseApp's underlying
// CommitMultiStore.
func (app *BaseApp) SetCommitMultiStoreTracer(w io.Writer) {
	app.cms.WithTracer(w)
}

// Mount a store to the provided key in the BaseApp multistore
func (app *BaseApp) mountStoresIAVL(keys ...*store.KVStoreKey) {
	for _, key := range keys {
		app.mountStore(key, store.StoreTypeIAVL)
	}
}

// Mount a store to the provided key in the BaseApp multistore, using the default DB
func (app *BaseApp) mountStore(key store.StoreKey, typ store.StoreType) {
	app.cms.MountStoreWithDB(key, typ, nil)
}

// load latest application version
func (app *BaseApp) LoadLatestVersion() error {
	err := app.cms.LoadLatestVersion()
	if err != nil {
		return err
	}
	return app.initFromStore()
}

// load application version
func (app *BaseApp) LoadVersion(version int64) error {
	err := app.cms.LoadVersion(version)
	if err != nil {
		return err
	}
	return app.initFromStore()
}

// the last CommitID of the multistore
func (app *BaseApp) LastCommitID() store.CommitID {
	return app.cms.LastCommitID()
}

// the last committed block height
func (app *BaseApp) LastBlockHeight() int64 {
	return app.cms.LastCommitID().Version
}

// initializes the remaining logic from app.cms
func (app *BaseApp) initFromStore() error {
	app.setCheckState(abci.Header{})
	app.Seal()
	return nil
}

// NewContext returns a new Context with the correct store, the given header, and nil txBytes.
func (app *BaseApp) NewContext(isCheckTx bool, header abci.Header) ctx.Context {
	if isCheckTx {
		return ctx.NewContext(app.checkState.ms, header, true, app.Logger, app.registerMappers)
	}
	return ctx.NewContext(app.deliverState.ms, header, false, app.Logger, app.registerMappers)
}

type state struct {
	ms  store.CacheMultiStore
	ctx ctx.Context
}

func (st *state) CacheMultiStore() store.CacheMultiStore {
	return st.ms.CacheMultiStore()
}

func (app *BaseApp) setCheckState(header abci.Header) {
	ms := app.cms.CacheMultiStore()
	app.checkState = &state{
		ms:  ms,
		ctx: ctx.NewContext(ms, header, true, app.Logger, app.registerMappers),
	}
}

func (app *BaseApp) setDeliverState(header abci.Header) {
	ms := app.cms.CacheMultiStore()
	app.deliverState = &state{
		ms:  ms,
		ctx: ctx.NewContext(ms, header, false, app.Logger, app.registerMappers),
	}

	//注入txQcpResultHandler
	app.deliverState.ctx = app.deliverState.ctx.WithTxQcpResultHandler(app.txQcpResultHandler)
}

//______________________________________________________________________________

// ABCI

// Implements ABCI
func (app *BaseApp) Info(req abci.RequestInfo) abci.ResponseInfo {
	lastCommitID := app.cms.LastCommitID()

	return abci.ResponseInfo{
		Data:             app.name,
		LastBlockHeight:  lastCommitID.Version,
		LastBlockAppHash: lastCommitID.Hash,
	}
}

// Implements ABCI
func (app *BaseApp) SetOption(req abci.RequestSetOption) (res abci.ResponseSetOption) {
	// TODO: Implement
	return
}

// Implements ABCI
// InitChain runs the initialization logic directly on the CommitMultiStore and commits it.
func (app *BaseApp) InitChain(req abci.RequestInitChain) (res abci.ResponseInitChain) {
	app.setDeliverState(abci.Header{ChainID: req.ChainId})
	app.setCheckState(abci.Header{ChainID: req.ChainId})

	if app.initChainer == nil {
		return
	}
	res = app.initChainer(app.deliverState.ctx, req)
	return
}

func splitPath(requestPath string) (path []string) {
	path = strings.Split(requestPath, "/")
	// first element is empty string
	if len(path) > 0 && path[0] == "" {
		path = path[1:]
	}
	return path
}

// Implements ABCI.
// Delegates to CommitMultiStore if it implements Queryable
func (app *BaseApp) Query(req abci.RequestQuery) (res abci.ResponseQuery) {
	path := splitPath(req.Path)
	if len(path) == 0 {
		msg := "no query path provided"
		return types.ErrUnknownRequest(msg).QueryResult()
	}
	switch path[0] {
	case "store":
		return handleQueryStore(app, path, req)
	}

	msg := "unknown query path"
	return types.ErrUnknownRequest(msg).QueryResult()
}

func handleQueryStore(app *BaseApp, path []string, req abci.RequestQuery) (res abci.ResponseQuery) {
	// "/store" prefix for store queries
	queryable, ok := app.cms.(store.Queryable)
	if !ok {
		msg := "multistore doesn't support queries"
		return types.ErrUnknownRequest(msg).QueryResult()
	}
	req.Path = "/" + strings.Join(path[1:], "/")
	return queryable.Query(req)
}

// BeginBlock implements the ABCI application interface.
func (app *BaseApp) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	if app.cms.TracingEnabled() {
		app.cms.ResetTraceContext()
		app.cms.WithTracingContext(store.TraceContext(
			map[string]interface{}{"blockHeight": req.Header.Height},
		))
	}

	// Initialize the DeliverTx state. If this is the first block, it should
	// already be initialized in InitChain. Otherwise app.deliverState will be
	// nil, since it is reset on Commit.
	if app.deliverState == nil {
		app.setDeliverState(req.Header)
	} else {
		// In the first block, app.deliverState.ctx will already be initialized
		// by InitChain. Context is now updated with Header information.
		app.deliverState.ctx = app.deliverState.ctx.WithBlockHeader(req.Header).WithBlockHeight(req.Header.Height)
	}

	//重置block tx index
	app.deliverState.ctx = app.deliverState.ctx.ResetBlockTxIndex()

	if app.beginBlocker != nil {
		res = app.beginBlocker(app.deliverState.ctx, req)
	}

	// set the signed validators for addition to context in deliverTx
	app.signedValidators = req.LastCommitInfo.GetValidators()
	return
}

// CheckTx implements ABCI
// CheckTx runs the "basic checks" to see whether or not a transaction can possibly be executed,
// first decoding, then the ante handler (which checks signatures/fees/ValidateBasic),
// then finally the route match to see whether a handler exists. CheckTx does not run the actual
// Msg handler function(s).
func (app *BaseApp) CheckTx(txBytes []byte) (res abci.ResponseCheckTx) {
	// Decode the Tx.
	var result types.Result
	var itx, err = app.txDecoder(txBytes)
	if err != nil {
		result = err.Result()
	} else {
		// 初始化context相关数据
		ctx := app.checkState.ctx.WithTxBytes(txBytes)
		switch tx := itx.(type) {
		case *txs.TxStd:
			result = app.checkTxStd(ctx, tx)
		case *txs.TxQcp:
			result = app.checkTxQcp(ctx, tx)
		default:
			result = types.ErrInternal("not support itx type").Result()
		}
	}

	return abci.ResponseCheckTx{
		Code:      uint32(result.Code),
		Data:      result.Data,
		Log:       result.Log,
		GasWanted: result.GasWanted,
		GasUsed:   result.GasUsed,
		Tags:      result.Tags,
	}
}

//checkTxStd: checkTx阶段对TxStd进行校验
func (app *BaseApp) checkTxStd(ctx ctx.Context, tx *txs.TxStd) (result types.Result) {

	defer func() {
		if r := recover(); r != nil {
			log := fmt.Sprintf("checkTxStd recovered : %v\nstack:\n%v", r, string(debug.Stack()))
			result = types.ErrInternal(log).Result()
		}
	}()

	//1. 校验txStd基础信息
	err := tx.ValidateBasicData(true)
	if err != nil {
		return err.Result()
	}

	//2. 校验签名
	_, res := app.ValidateTxStdUserSignatureAndNonce(ctx, tx)
	if !res.IsOK() {
		return res
	}

	return
}

func (app *BaseApp) ValidateTxStdUserSignatureAndNonce(ctx ctx.Context, tx *txs.TxStd) (newctx ctx.Context, result types.Result) {
	// TODO 细化操作
	//accountMapper 未设置时， 不做签名校验
	if getAccountMapper(ctx) == nil {
		app.Logger.Info("accountMapper not setup....")
		return
	}

	return
}

//checkTxQcp: checkTx阶段对TxQcp进行校验
func (app *BaseApp) checkTxQcp(ctx ctx.Context, tx *txs.TxQcp) (result types.Result) {

	defer func() {
		if r := recover(); r != nil {
			log := fmt.Sprintf("checkTxQcp recovered: %v\nstack:\n%v", r, string(debug.Stack()))
			result = types.ErrInternal(log).Result()
		}
	}()

	//1. 校验txQcp基础数据
	err := tx.ValidateBasicData(true, ctx.ChainID())
	if err != nil {
		return err.Result()
	}
	//2. 校验txQcp sequence: sequence 必须大于当前接收的该链最大的sequence
	//checkTx时仅校验 sequence > maxReceivedSeq
	maxReceivedSeq := getQcpMapper(ctx).GetMaxChainInSequence(tx.From)
	if tx.Sequence < maxReceivedSeq {
		return types.ErrInvalidSequence("tx sequence is less then received sequence").Result()
	}

	//3. 校验签名
	res := app.validateTxQcpSignature(ctx, tx)
	if !res.IsOK() {
		return res
	}

	return
}

//校验QcpTx签名是否正确
func (app *BaseApp) validateTxQcpSignature(ctx ctx.Context, qcpTx *txs.TxQcp) (result types.Result) {
	//1. 校验qcpTx签名者PubKey是否合法: TODO 需要区块链实现
	pubkey := qcpTx.Sig.Pubkey
	truestPubkey := getQcpMapper(ctx).GetChainInTruestPubKey(qcpTx.From)

	if truestPubkey != nil && pubkey != nil && bytes.Compare(pubkey.Bytes(), truestPubkey.Bytes()) != 0 {
		return types.ErrInvalidPubKey("qcpTx's signer is not valid").Result()
	}

	//2. 校验签名是否合法
	signedBytes := qcpTx.GetSigData()
	if !pubkey.VerifyBytes(signedBytes, qcpTx.Sig.Signature) {
		return types.ErrUnauthorized("signature verification failed").Result()
	}

	return
}

// Implements ABCI
func (app *BaseApp) DeliverTx(txBytes []byte) (res abci.ResponseDeliverTx) {

	//deliverTx处理tx时，设置tx index
	lastBlockTxIndex := app.deliverState.ctx.BlockTxIndex()
	app.deliverState.ctx = app.deliverState.ctx.WithBlockTxIndex(lastBlockTxIndex + 1)

	// Decode the Tx.
	var result types.Result
	var itx, err = app.txDecoder(txBytes)
	if err != nil {
		result = err.Result()
		return abci.ResponseDeliverTx{
			Code:      uint32(result.Code),
			Data:      result.Data,
			Log:       result.Log,
			GasWanted: result.GasWanted,
			GasUsed:   result.GasUsed,
			Tags:      result.Tags,
		}
	}

	//初始化context相关数据
	ctx := app.deliverState.ctx.WithTxBytes(txBytes).WithSigningValidators(app.signedValidators)

	isTxStd := false
	isTxQcpResult := false
	originSequence := int64(0)
	originFrom := ""

	var crossTxQcp *txs.TxQcp

	switch tx := itx.(type) {
	case *txs.TxStd:
		isTxStd = true
		result, crossTxQcp = app.deliverTxStd(ctx, tx)
	case *txs.TxQcp:
		if tx.IsResult {
			isTxQcpResult = true
		}
		originSequence = tx.Sequence
		originFrom = tx.From
		result, crossTxQcp = app.deliverTxQcp(ctx, tx)
	default:
		result = types.ErrInternal("not support itx type").Result()
	}

	if isTxStd {
		// crossTxQcp 不为空时，需要将跨链结果保存
		if crossTxQcp != nil && app.signerForCrossTxQcp == nil {
			app.Logger.Error("exsits cross txqcp, but signer is nil.if you forgot to set up signer?")
			return
		}
		if crossTxQcp != nil {
			txQcp := getQcpMapper(ctx).SaveCrossChainResult(ctx, crossTxQcp.Payload, crossTxQcp.To, false, app.signerForCrossTxQcp)
			result.Tags = result.Tags.AppendTag("qcp.From", []byte(txQcp.From)).
				AppendTag("qcp.To", []byte(txQcp.To)).
				AppendTag("qcp.Sequence", types.Int2Byte(txQcp.Sequence))
			// .AppendTag("qcp.HashBytes",)
		}
	}

	if !isTxStd && !isTxQcpResult {
		//类型为TxQcp时，将所有结果进行保存
		txQcpResult := &txs.QcpTxResult{
			Code:                int64(result.Code),
			Extends:             make([]cmn.KVPair, 5),
			GasUsed:             types.NewInt(result.GasUsed),
			QcpOriginalSequence: originSequence,
			Info:                result.Log,
		}

		result.Tags = result.Tags.AppendTag("From", []byte(ctx.ChainID())).
			AppendTag("To", []byte(originFrom))

		txQcpResult.Extends = append(txQcpResult.Extends, result.Tags...)

		payload := txs.TxStd{
			ITx:       txQcpResult,
			Signature: make([]txs.Signature, 0),
			ChainID:   ctx.ChainID(),
			MaxGas:    types.ZeroInt(),
		}

		txQcp := getQcpMapper(ctx).SaveCrossChainResult(ctx, payload, originFrom, true, nil)
		result.Tags = result.Tags.AppendTag("Sequence", types.Int2Byte(txQcp.Sequence))
		// .AppendTag("HashBytes",)
	}

	// Tell the blockchain engine (i.e. Tendermint).
	return abci.ResponseDeliverTx{
		Code:      uint32(result.Code),
		Data:      result.Data,
		Log:       result.Log,
		GasWanted: result.GasWanted,
		GasUsed:   result.GasUsed,
		Tags:      result.Tags,
	}
}

//deliverTxStd: deliverTx阶段对TxStd进行业务处理
func (app *BaseApp) deliverTxStd(ctx ctx.Context, tx *txs.TxStd) (result types.Result, crossTxQcp *txs.TxQcp) {

	defer func() {
		if r := recover(); r != nil {
			log := fmt.Sprintf("deliverTxStd recovered: %v\nstack:\n%v", r, string(debug.Stack()))
			result = types.ErrInternal(log).Result()
		}
	}()

	//1. 校验基础数据
	err := tx.ValidateBasicData(false)
	if err != nil {
		return err.Result(), nil
	}
	//2. 校验签名
	newctx, res := app.ValidateTxStdUserSignatureAndNonce(ctx, tx)
	if !res.IsOK() {
		return res, nil
	}

	if !newctx.IsZero() {
		ctx = newctx
	}

	//3. 执行exec: 需要开启临时缓存存储状态
	msCache := getState(app, false).CacheMultiStore()
	if msCache.TracingEnabled() {
		msCache = msCache.WithTracingContext(store.TraceContext(
			map[string]interface{}{"txHash": cmn.HexBytes(tmhash.Sum(ctx.TxBytes())).String()},
		)).(store.CacheMultiStore)
	}

	ctx = ctx.WithMultiStore(msCache)
	result, crossTxQcp = tx.ITx.Exec(ctx)

	if result.IsOK() {
		msCache.Write()
	}

	return
}

//deliverTxQcp: devilerTx阶段对TxQcp进行业务处理
func (app *BaseApp) deliverTxQcp(ctx ctx.Context, tx *txs.TxQcp) (result types.Result, crossTxQcp *txs.TxQcp) {

	defer func() {
		if r := recover(); r != nil {
			log := fmt.Sprintf("deliverTxQcp recovered: %v\nstack:\n%v", r, string(debug.Stack()))
			result = types.ErrInternal(log).Result()
		}
	}()

	//1. 校验TxQcp基础数据
	err := tx.ValidateBasicData(false, ctx.ChainID())
	if err != nil {
		return err.Result(), nil
	}

	//2. 校验TxQcp sequence: sequence = maxInSequence + 1
	// deliverTx时校验 sequence =  maxInSequence + 1
	maxInSequence := getQcpMapper(ctx).GetMaxChainInSequence(tx.From)
	if tx.Sequence != maxInSequence+1 {
		result = types.ErrInvalidSequence("tx sequence is less then received sequence").Result()
		return
	}

	//3. 更新qcp in sequence
	getQcpMapper(ctx).SetMaxChainInSequence(tx.From, maxInSequence+1)

	//4. 校验TxQcp签名
	res := app.validateTxQcpSignature(ctx, tx)
	if !res.IsOK() {
		result = res
		return
	}

	result, crossTxQcp = app.deliverTxStd(ctx, &tx.Payload)
	return
}

// Returns the applicantion's deliverState
func getState(app *BaseApp, isCheckTx bool) *state {
	if isCheckTx {
		return app.checkState
	}

	return app.deliverState
}

// EndBlock implements the ABCI application interface.
func (app *BaseApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	if app.deliverState.ms.TracingEnabled() {
		app.deliverState.ms = app.deliverState.ms.ResetTraceContext().(store.CacheMultiStore)
	}

	if app.endBlocker != nil {
		res = app.endBlocker(app.deliverState.ctx, req)
	}

	return
}

// Implements ABCI
func (app *BaseApp) Commit() (res abci.ResponseCommit) {
	header := app.deliverState.ctx.BlockHeader()

	// Write the Deliver state and commit the MultiStore
	app.deliverState.ms.Write()
	commitID := app.cms.Commit()
	// TODO: this is missing a module identifier and dumps byte array
	app.Logger.Debug("Commit synced",
		"commit", commitID,
	)

	// Reset the Check state to the latest committed
	// NOTE: safe because Tendermint holds a lock on the mempool for Commit.
	// Use the header from this latest block.
	app.setCheckState(header)

	// Empty the Deliver state
	app.deliverState = nil

	return abci.ResponseCommit{
		Data: commitID.Hash,
	}
}

//TODO: 待优化
func getQcpMapper(ctx ctx.Context) *qcp.QcpMapper {
	return ctx.Mapper(qcp.QcpMapperName).(*qcp.QcpMapper)
}

//TODO: 待优化
func getAccountMapper(ctx ctx.Context) *account.AccountMapper {
	return ctx.Mapper(account.AccountMapperName).(*account.AccountMapper)
}
