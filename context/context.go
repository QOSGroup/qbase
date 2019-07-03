// nolint
package context

import (
	"context"
	"sync"
	"time"

	"github.com/QOSGroup/qbase/store/gaskv"
	storetypes "github.com/QOSGroup/qbase/store/types"

	"github.com/QOSGroup/qbase/mapper"

	"github.com/QOSGroup/qbase/store"
	"github.com/golang/protobuf/proto"

	"github.com/QOSGroup/qbase/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
)

/*
The intent of Context is for it to be an immutable object that can be
cloned and updated cheaply with WithValue() and passed forward to the
next decorator or handler. For example,

 func MsgHandler(ctx Context, tx Tx) Result {
 	...
 	ctx = ctx.WithValue(key, value)
 	...
 }
*/
type Context struct {
	context.Context
	pst *thePast
	gen int
	// Don't add any other fields here,
	// it's probably not what you want to do.
}

// create a new context
// nolint: unparam
func NewContext(ms store.MultiStore, header abci.Header, isCheckTx bool, logger log.Logger, registSeedMapper map[string]mapper.IMapper) Context {
	c := Context{
		Context: context.Background(),
		pst:     newThePast(),
		gen:     0,
	}
	c = c.WithGasMeter(types.NewInfiniteGasMeter())
	c = c.WithMultiStore(ms)
	c = c.WithBlockHeader(header)
	c = c.WithBlockHeight(header.Height)
	c = c.WithChainID(header.ChainID)
	c = c.WithIsCheckTx(isCheckTx)
	c = c.WithTxBytes(nil)
	c = c.WithLogger(logger)
	c = c.WithVoteInfos(nil)
	c = c.WithConsensusParams(nil)
	c = c.WithMinimumFees([]types.Coin{&types.BaseCoin{}})
	c = c.withRegisteredMap(registSeedMapper)
	c = c.copyKVStoreMapperFromSeed()
	return c
}

// is context nil
func (c Context) IsZero() bool {
	return c.Context == nil
}

//----------------------------------------
// Getting a value

// context value for the provided key
func (c Context) Value(key interface{}) interface{} {
	value := c.Context.Value(key)
	if cloner, ok := value.(cloner); ok {
		return cloner.Clone()
	}
	if message, ok := value.(proto.Message); ok {
		return proto.Clone(message)
	}
	return value
}

// KVStore fetches a KVStore from the MultiStore.
func (c Context) KVStore(key store.StoreKey) store.KVStore {
	return c.MultiStore().GetKVStore(key)
}

// TransientStore fetches a TransientStore from the MultiStore.
func (c Context) TransientStore(key store.StoreKey) store.KVStore {
	return c.MultiStore().GetKVStore(key)
}

//----------------------------------------
// With* (setting a value)

// nolint
func (c Context) WithValue(key interface{}, value interface{}) Context {
	return c.withValue(key, value)
}
func (c Context) WithCloner(key interface{}, value cloner) Context {
	return c.withValue(key, value)
}
func (c Context) WithCacheWrapper(key interface{}, value store.CacheWrapper) Context {
	return c.withValue(key, value)
}
func (c Context) WithProtoMsg(key interface{}, value proto.Message) Context {
	return c.withValue(key, value)
}
func (c Context) WithString(key interface{}, value string) Context {
	return c.withValue(key, value)
}
func (c Context) WithInt32(key interface{}, value int32) Context {
	return c.withValue(key, value)
}
func (c Context) WithUint32(key interface{}, value uint32) Context {
	return c.withValue(key, value)
}
func (c Context) WithUint64(key interface{}, value uint64) Context {
	return c.withValue(key, value)
}

func (c Context) Mapper(name string) mapper.IMapper {
	registeredMapper := c.Value(contextKeyCurrentRegisteredMapper).(map[string]mapper.IMapper)
	if mapper, ok := registeredMapper[name]; ok {
		return mapper
	}
	return nil
}

func (c Context) withValue(key interface{}, value interface{}) Context {
	c.pst.bump(Op{
		gen:   c.gen + 1,
		key:   key,
		value: value,
	}) // increment version for all relatives.

	return Context{
		Context: context.WithValue(c.Context, key, value),
		pst:     c.pst,
		gen:     c.gen + 1,
	}
}

//----------------------------------------
// Values that require no key.

type contextKey int // local to the context module

const (
	contextKeyMultiStore contextKey = iota
	contextKeyBlockHeader
	contextKeyBlockHeight
	contextKeyConsensusParams
	contextKeyChainID
	contextKeyIsCheckTx
	contextKeyTxBytes
	contextKeyLogger
	contextKeyVoteInfos
	contextKeyGasMeter
	contextKeyBlockGasMeter
	contextKeyMinimumFees
	//增加特定的context key
	contextKeyBlockTxIndex       // tx在block中的索引
	contextKeyTxQcpResultHandler //处理TxQcpResult回调函数
	contextKeyRegisteredMapper   //注册的mapper
	contextKeyCurrentRegisteredMapper
	contextKeyEventManager
)

//ContextKeySigners 用于保存tx中签名的账户
const ContextKeySigners contextKey = 99999

// NOTE: Do not expose MultiStore.
// MultiStore exposes all the keys.
// Instead, pass the context and the store key.
func (c Context) MultiStore() store.MultiStore {
	return c.Value(contextKeyMultiStore).(store.MultiStore)
}

func (c Context) BlockHeader() abci.Header { return c.Value(contextKeyBlockHeader).(abci.Header) }

func (c Context) BlockHeight() int64 { return c.Value(contextKeyBlockHeight).(int64) }

func (c Context) ConsensusParams() *abci.ConsensusParams {
	return c.Value(contextKeyConsensusParams).(*abci.ConsensusParams)
}

func (c Context) VoteInfos() []abci.VoteInfo {
	return c.Value(contextKeyVoteInfos).([]abci.VoteInfo)
}

func (c Context) ChainID() string { return c.Value(contextKeyChainID).(string) }

func (c Context) TxBytes() []byte { return c.Value(contextKeyTxBytes).([]byte) }

func (c Context) Logger() log.Logger { return c.Value(contextKeyLogger).(log.Logger) }

func (c Context) GasMeter() types.GasMeter { return c.Value(contextKeyGasMeter).(types.GasMeter) }

func (c Context) IsCheckTx() bool { return c.Value(contextKeyIsCheckTx).(bool) }

func (c Context) MinimumFees() []types.Coin { return c.Value(contextKeyMinimumFees).([]types.Coin) }

func (c Context) BlockTxIndex() int64 {
	index := c.Value(contextKeyBlockTxIndex)
	if index == nil {
		return int64(-1)
	}

	return index.(int64)
}

func (c Context) TxQcpResultHandler() func(ctx Context, itx interface{}) {
	return c.Value(contextKeyTxQcpResultHandler).(func(ctx Context, itx interface{}))
}

func (c Context) WithMultiStore(ms store.MultiStore) Context {
	newCtx := c.withValue(contextKeyMultiStore, ms)
	return newCtx.copyKVStoreMapperFromSeed()
}

func (c Context) WithBlockHeader(header abci.Header) Context {
	var _ proto.Message = &header // for cloning.
	return c.withValue(contextKeyBlockHeader, header)
}

func (c Context) WithBlockTime(newTime time.Time) Context {
	newHeader := c.BlockHeader()
	newHeader.Time = newTime
	return c.WithBlockHeader(newHeader)
}

func (c Context) WithProposer(addr types.Address) Context {
	newHeader := c.BlockHeader()
	newHeader.ProposerAddress = addr.Bytes()
	return c.WithBlockHeader(newHeader)
}

func (c Context) WithBlockHeight(height int64) Context {
	newHeader := c.BlockHeader()
	newHeader.Height = height
	return c.withValue(contextKeyBlockHeight, height).withValue(contextKeyBlockHeader, newHeader)
}
func (c Context) WithConsensusParams(params *abci.ConsensusParams) Context {
	if params == nil {
		return c
	}
	return c.withValue(contextKeyConsensusParams, params)
}

func (c Context) WithChainID(chainID string) Context { return c.withValue(contextKeyChainID, chainID) }

func (c Context) WithTxBytes(txBytes []byte) Context { return c.withValue(contextKeyTxBytes, txBytes) }

func (c Context) WithLogger(logger log.Logger) Context { return c.withValue(contextKeyLogger, logger) }

func (c Context) WithVoteInfos(VoteInfos []abci.VoteInfo) Context {
	return c.withValue(contextKeyVoteInfos, VoteInfos)
}

//
func (c Context) withRegisteredMap(registeSeedMapper map[string]mapper.IMapper) Context {
	return c.withValue(contextKeyRegisteredMapper, registeSeedMapper)
}

//mapper与store有关，当store变化时，需要创建基于当前store的mapper
func (c Context) copyKVStoreMapperFromSeed() Context {
	mapperWithStore := make(map[string]mapper.IMapper)

	v := c.Value(contextKeyRegisteredMapper)
	if v == nil {
		return c
	}

	registeredMapper := v.(map[string]mapper.IMapper)
	if len(registeredMapper) > 0 {
		for name, mapper := range registeredMapper {
			cpyMapper := mapper.Copy()
			store := gaskv.NewStore(c.KVStore(mapper.GetStoreKey()), c.GasMeter(), storetypes.KVGasConfig())
			cpyMapper.SetStore(store)
			mapperWithStore[name] = cpyMapper
		}
	}

	return c.withValue(contextKeyCurrentRegisteredMapper, mapperWithStore)
}

func (c Context) WithGasMeter(meter types.GasMeter) Context {
	var ctx = c.withValue(contextKeyGasMeter, meter)
	return ctx.copyKVStoreMapperFromSeed()
}

func (c Context) WithBlockGasMeter(meter types.GasMeter) Context {
	return c.withValue(contextKeyBlockGasMeter, meter)
}

func (c Context) WithIsCheckTx(isCheckTx bool) Context {
	return c.withValue(contextKeyIsCheckTx, isCheckTx)
}

func (c Context) WithMinimumFees(minFees []types.Coin) Context {
	return c.withValue(contextKeyMinimumFees, minFees)
}

func (c Context) WithBlockTxIndex(blockTxIndex int64) Context {
	return c.withValue(contextKeyBlockTxIndex, blockTxIndex)
}

func (c Context) WithTxQcpResultHandler(txQcpResultHandler func(ctx Context, itx interface{})) Context {
	return c.withValue(contextKeyTxQcpResultHandler, txQcpResultHandler)
}

func (c Context) ResetBlockTxIndex() Context {
	return c.withValue(contextKeyBlockTxIndex, int64(-1))
}

func (c Context) WithEventManager(em *types.EventManager) Context {
	return c.WithValue(contextKeyEventManager, em)
}

func (c Context) EventManager() *types.EventManager { return c.Value(contextKeyEventManager).(*types.EventManager) }

// Cache the multistore and return a new cached context. The cached context is
// written to the context when writeCache is called.
func (c Context) CacheContext() (cc Context, writeCache func()) {
	cms := c.MultiStore().CacheMultiStore()
	cc = c.WithMultiStore(cms)
	return cc, cms.Write
}

//----------------------------------------
// thePast

// Returns false if ver <= 0 || ver > len(c.pst.ops).
// The first operation is version 1.
func (c Context) GetOp(ver int64) (Op, bool) {
	return c.pst.getOp(ver)
}

//----------------------------------------
// Misc.

type cloner interface {
	Clone() interface{} // deep copy
}

// XXX add description
type Op struct {
	// type is always 'with'
	gen   int
	key   interface{}
	value interface{}
}

type thePast struct {
	mtx sync.RWMutex
	ver int
	ops []Op
}

func newThePast() *thePast {
	return &thePast{
		ver: 0,
		ops: nil,
	}
}

func (pst *thePast) bump(op Op) {
	pst.mtx.Lock()
	pst.ver++
	pst.ops = append(pst.ops, op)
	pst.mtx.Unlock()
}

func (pst *thePast) version() int {
	pst.mtx.RLock()
	defer pst.mtx.RUnlock()
	return pst.ver
}

// Returns false if ver <= 0 || ver > len(pst.ops).
// The first operation is version 1.
func (pst *thePast) getOp(ver int64) (Op, bool) {
	pst.mtx.RLock()
	defer pst.mtx.RUnlock()
	l := int64(len(pst.ops))
	if l < ver || ver <= 0 {
		return Op{}, false
	}
	return pst.ops[ver-1], true
}
