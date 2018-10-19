package baseabci

import (
	"bytes"
	"fmt"
	"io"
	"runtime/debug"
	"strings"

	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/mapper"
	"github.com/QOSGroup/qbase/qcp"
	"github.com/QOSGroup/qbase/version"

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
	Logger log.Logger
	name   string                 // application name from abci.Info
	db     dbm.DB                 // common DB backend
	cms    store.CommitMultiStore // Main (uncached) state

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
	txQcpResultHandler TxQcpResultHandler // exec方法中回调，执行具体的业务逻辑
	txQcpSigner        crypto.PrivKey     //对跨链TxQcp签名的私钥， 由app在启动时初始化

	//注册自定义查询处理
	customQueryHandler CustomQueryHandler
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

	app := &BaseApp{
		Logger:          logger,
		name:            name,
		db:              db,
		cms:             store.NewCommitMultiStore(db),
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
	case "app":
		return handleQueryApp(app, path, req)
	case "store":
		return handleQueryStore(app, path, req)
	case "custom":
		return handlerCustomQuery(app, path, req)
	}

	msg := "unknown query path"
	return types.ErrUnknownRequest(msg).QueryResult()
}

func handleQueryApp(app *BaseApp, path []string, req abci.RequestQuery) (res abci.ResponseQuery) {
	if len(path) >= 2 {
		var result interface{}
		switch path[1] {
		case "version":
			result = version.Version
		default:
			result = types.ErrUnknownRequest(fmt.Sprintf("Unknown query: %s", path)).Result()
		}

		value := app.cdc.MustMarshalBinaryBare(result)
		return abci.ResponseQuery{
			Code:  uint32(types.ABCICodeOK),
			Value: value,
		}
	}
	msg := "Expected second parameter to be either simulate or version, neither was present"
	return types.ErrUnknownRequest(msg).QueryResult()
}

func handlerCustomQuery(app *BaseApp, path []string, req abci.RequestQuery) (res abci.ResponseQuery) {

	if app.customQueryHandler == nil {
		return types.ErrUnknownRequest("custom query handler not register").QueryResult()
	}

	ctx := ctx.NewContext(app.cms.CacheMultiStore(), app.checkState.ctx.BlockHeader(), true, app.Logger, app.registerMappers)
	bz, err := app.customQueryHandler(ctx, path[1:], req)

	if err != nil {
		return abci.ResponseQuery{
			Code: uint32(err.ABCICode()),
			Log:  err.ABCILog(),
		}
	}

	return abci.ResponseQuery{
		Code:  uint32(types.ABCICodeOK),
		Value: bz,
	}

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
	var tx, err = types.DecoderTx(app.cdc, txBytes)

	if err != nil {
		return toResponseCheckTx(err.Result())
	}

	// 初始化context相关数据
	ctx := app.checkState.ctx.WithTxBytes(txBytes)
	switch implTx := tx.(type) {
	case *txs.TxStd:
		result = app.checkTxStd(ctx, implTx)
	case *txs.TxQcp:
		result = app.checkTxQcp(ctx, implTx)
	default:
		result = types.ErrInternal("not support itx type").Result()
	}

	return toResponseCheckTx(result)
}

func toResponseCheckTx(result types.Result) abci.ResponseCheckTx {
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
	err := tx.ValidateBasicData(ctx, true, ctx.ChainID())
	if err != nil {
		return err.Result()
	}

	//2. 校验签名
	_, res := app.validateTxStdUserSignatureAndNonce(ctx, tx)
	if !res.IsOK() {
		return res
	}

	return
}

//校验txstd用户签名，签名通过后，增加用户none
func (app *BaseApp) validateTxStdUserSignatureAndNonce(cctx ctx.Context, tx *txs.TxStd) (newctx ctx.Context, result types.Result) {
	//accountMapper 未设置时， 不做签名校验
	accounMapper := getAccountMapper(cctx)

	if accounMapper == nil {
		app.Logger.Info("accountMapper not setup....")
		return
	}

	//签名者为空则不校验签名
	signers := tx.ITx.GetSigner()
	if len(signers) == 0 {
		return
	}

	signatures := tx.Signature
	signerAccount := make([]account.Account, len(signers))
	for i, addr := range signers {
		acc := accounMapper.GetAccount(addr)
		if acc == nil {
			acc = accounMapper.NewAccountWithAddress(addr)
		}
		signerAccount[i] = acc
	}

	//校验account address,nonce是否与签名中一致. 并设置account pubkey
	for i := 0; i < len(signatures); i++ {
		acc := signerAccount[i]
		signature := signatures[i]

		if signature.Pubkey != nil {
			pubkeyAddress := types.Address(signature.Pubkey.Address())
			if !bytes.Equal(pubkeyAddress, acc.GetAddress()) {
				result = types.ErrInternal(fmt.Sprintf("invalid address. expect: %s, got: %s", acc.GetAddress(), pubkeyAddress)).Result()
				return
			}
		}

		if uint64(signature.Nonce) != acc.GetNonce() {
			result = types.ErrInternal(fmt.Sprintf("invalid nonce. expect: %d, got: %d", acc.GetNonce(), signature.Nonce)).Result()
			return
		}

		if acc.GetPubicKey() == nil {
			if signature.Pubkey == nil {
				result = types.ErrInternal(fmt.Sprintf("signature missing pubkey")).Result()
				return
			}
			acc.SetPublicKey(signature.Pubkey)
		}
	}

	//校验签名并增加账户nonce
	for i := 0; i < len(signatures); i++ {
		acc := signerAccount[i]
		signature := signatures[i]
		pubkey := acc.GetPubicKey()
		//1. 根据账户nonce及tx生成signData
		signBytes := append(tx.GetSignData(), types.Int2Byte(int64(acc.GetNonce()))...)
		if !pubkey.VerifyBytes(signBytes, signature.Signature) {
			result = types.ErrInternal(fmt.Sprintf("signature verification failed")).Result()
			return
		}

		//acccount nonce increment
		acc.SetNonce(acc.GetNonce() + 1)
		accounMapper.SetAccount(acc)

		signerAccount[i] = acc
	}

	newctx = cctx.WithValue(ctx.ContextKeySigners, signerAccount)
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
		return types.ErrInvalidSequence(fmt.Sprintf("tx sequence is less then received sequence. max is %d , got is %d", maxReceivedSeq, tx.Sequence)).Result()
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
	//1. 校验qcpTx签名者PubKey是否合法:
	pubkey := qcpTx.Sig.Pubkey
	trustPubkey := getQcpMapper(ctx).GetChainInTruestPubKey(qcpTx.From)

	if trustPubkey == nil {
		return types.ErrInvalidPubKey("trust pubkey not set. you should set one trustKey per chain").Result()
	}

	if pubkey == nil {
		return types.ErrInvalidPubKey("pubkey is nil in signature").Result()
	}

	if !bytes.Equal(pubkey.Bytes(), trustPubkey.Bytes()) {
		return types.ErrInvalidPubKey(fmt.Sprintf("pubkey is signature is not expect. Got: %X , Expect: %X", pubkey.Bytes(), trustPubkey.Bytes())).Result()
	}

	//2. 校验签名是否合法
	sigBytes := qcpTx.GetSigData()
	if !pubkey.VerifyBytes(sigBytes, qcpTx.Sig.Signature) {
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
	var tx, err = types.DecoderTx(app.cdc, txBytes)
	if err != nil {
		result = err.Result()
		return toResponseDeliverTx(result)
	}

	//初始化context相关数据
	ctx := app.deliverState.ctx.WithTxBytes(txBytes).WithSigningValidators(app.signedValidators)

	switch implTx := tx.(type) {
	case *txs.TxStd:
		result = app.deliverTxStd(ctx, implTx)
	case *txs.TxQcp:
		result = app.deliverTxQcp(ctx, implTx)
	default:
		result = types.ErrInternal("not support itx type").Result()
	}

	// Tell the blockchain engine (i.e. Tendermint).
	return toResponseDeliverTx(result)
}

func toResponseDeliverTx(result types.Result) abci.ResponseDeliverTx {
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
func (app *BaseApp) deliverTxStd(ctx ctx.Context, tx *txs.TxStd) (result types.Result) {

	defer func() {
		if r := recover(); r != nil {
			log := fmt.Sprintf("deliverTxStd recovered: %v\nstack:\n%v", r, string(debug.Stack()))
			result = types.ErrInternal(log).Result()
		}
	}()

	//1. 校验基础数据
	err := tx.ValidateBasicData(ctx, false, ctx.ChainID())
	if err != nil {
		return err.Result()
	}
	//2. 校验签名
	newctx, res := app.validateTxStdUserSignatureAndNonce(ctx, tx)
	if !res.IsOK() {
		return res
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

	var crossTxQcp *txs.TxQcp
	result, crossTxQcp = tx.ITx.Exec(ctx)

	//4. 根据crossTxQcp结果判断是否保存跨链结果
	// crossTxQcp 不为空时，需要将跨链结果保存
	if crossTxQcp != nil && app.txQcpSigner == nil {
		app.Logger.Error("exsits cross txqcp, but signer is nil.if you forgot to set up signer?")
	}

	if crossTxQcp != nil {
		txQcp := getQcpMapper(ctx).SaveCrossChainResult(ctx, crossTxQcp.TxStd, crossTxQcp.To, false, app.txQcpSigner)
		result.Tags = result.Tags.AppendTag(qcp.QcpFrom, []byte(txQcp.From)).
			AppendTag(qcp.QcpTo, []byte(txQcp.To)).
			AppendTag(qcp.QcpSequence, types.Int2Byte(txQcp.Sequence)).
			AppendTag(qcp.QcpHashBytes, crypto.Sha256(txQcp.GetSigData()))
	}

	if result.IsOK() {
		msCache.Write()
	}

	return
}

//deliverTxQcp: devilerTx阶段对TxQcp进行业务处理
func (app *BaseApp) deliverTxQcp(ctx ctx.Context, tx *txs.TxQcp) (result types.Result) {

	defer func() {
		if r := recover(); r != nil {
			log := fmt.Sprintf("deliverTxQcp recovered: %v\nstack:\n%v", r, string(debug.Stack()))
			result = types.ErrInternal(log).Result()
		}

		//6. txQcp不为result时， 保存执行结果
		if !tx.IsResult {
			//类型为TxQcp时，将所有结果进行保存
			txQcpResult := &txs.QcpTxResult{
				Code:                int64(result.Code),
				Extends:             make([]cmn.KVPair, 1),
				GasUsed:             types.NewInt(result.GasUsed),
				QcpOriginalSequence: tx.Sequence,
				Info:                result.Log,
			}

			result.Tags = result.Tags.AppendTag(qcp.QcpFrom, []byte(ctx.ChainID())).
				AppendTag(qcp.QcpTo, []byte(tx.From))

			txQcpResult.Extends = append(txQcpResult.Extends, result.Tags...)

			txStd := &txs.TxStd{
				ITx:       txQcpResult,
				Signature: make([]txs.Signature, 0),
				ChainID:   ctx.ChainID(),
				MaxGas:    types.ZeroInt(),
			}

			txQcp := getQcpMapper(ctx).SaveCrossChainResult(ctx, txStd, tx.From, true, nil)
			result.Tags = result.Tags.AppendTag(qcp.QcpSequence, types.Int2Byte(txQcp.Sequence)).
				AppendTag(qcp.QcpHashBytes, crypto.Sha256(txQcp.GetSigData()))
		}

	}()

	//1. 校验TxQcp基础数据
	err := tx.ValidateBasicData(false, ctx.ChainID())
	if err != nil {
		result = err.Result()
		return
	}

	//2. 校验TxQcp sequence: sequence = maxInSequence + 1
	// deliverTx时校验 sequence =  maxInSequence + 1
	maxInSequence := getQcpMapper(ctx).GetMaxChainInSequence(tx.From)
	if tx.Sequence != maxInSequence+1 {
		result = types.ErrInvalidSequence(fmt.Sprintf("tx Sequence is not equals maxInSequence + 1 . maxInSequence is %d , tx.Sequence is %d", maxInSequence, tx.Sequence)).Result()
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

	//5. 执行内部txStd
	result = app.deliverTxStd(ctx, tx.TxStd)
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

func getQcpMapper(ctx ctx.Context) *qcp.QcpMapper {
	mapper := ctx.Mapper(qcp.QcpMapperName)
	if mapper == nil {
		return nil
	}
	return mapper.(*qcp.QcpMapper)
}

func getAccountMapper(ctx ctx.Context) *account.AccountMapper {
	mapper := ctx.Mapper(account.AccountMapperName)
	if mapper == nil {
		return nil
	}
	return mapper.(*account.AccountMapper)
}
