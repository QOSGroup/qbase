package context_test

import (
	"fmt"
	"testing"

	"github.com/QOSGroup/qbase/context"
	"github.com/QOSGroup/qbase/store"
	"github.com/QOSGroup/qbase/types"
	"github.com/stretchr/testify/require"

	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	abci "github.com/tendermint/tendermint/abci/types"
)

type MockLogger struct {
	logs *[]string
}

func NewMockLogger() MockLogger {
	logs := make([]string, 0)
	return MockLogger{
		&logs,
	}
}

func (l MockLogger) Debug(msg string, kvs ...interface{}) {
	*l.logs = append(*l.logs, msg)
}

func (l MockLogger) Info(msg string, kvs ...interface{}) {
	*l.logs = append(*l.logs, msg)
}

func (l MockLogger) Error(msg string, kvs ...interface{}) {
	*l.logs = append(*l.logs, msg)
}

func (l MockLogger) With(kvs ...interface{}) log.Logger {
	panic("not implemented")
}

func TestContextGetOpShouldNeverPanic(t *testing.T) {
	var ms store.MultiStore
	ctx := context.NewContext(ms, abci.Header{}, false, log.NewNopLogger(), nil)
	indices := []int64{
		-10, 1, 0, 10, 20,
	}

	for _, index := range indices {
		_, _ = ctx.GetOp(index)
	}
}

func defaultContext(key store.StoreKey) context.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, store.StoreTypeIAVL, db)
	cms.LoadLatestVersion()
	ctx := context.NewContext(cms, abci.Header{}, false, log.NewNopLogger(), nil)
	return ctx
}

func TestCacheContext(t *testing.T) {
	key := store.NewKVStoreKey(t.Name())
	k1 := []byte("hello")
	v1 := []byte("world")
	k2 := []byte("key")
	v2 := []byte("value")

	ctx := defaultContext(key)
	store := ctx.KVStore(key)
	store.Set(k1, v1)
	require.Equal(t, v1, store.Get(k1))
	require.Nil(t, store.Get(k2))

	cctx, write := ctx.CacheContext()
	cstore := cctx.KVStore(key)
	require.Equal(t, v1, cstore.Get(k1))
	require.Nil(t, cstore.Get(k2))

	cstore.Set(k2, v2)
	require.Equal(t, v2, cstore.Get(k2))
	require.Nil(t, store.Get(k2))

	write()

	require.Equal(t, v2, store.Get(k2))
}

func TestLogContext(t *testing.T) {
	key := store.NewKVStoreKey(t.Name())
	ctx := defaultContext(key)
	logger := NewMockLogger()
	ctx = ctx.WithLogger(logger)
	ctx.Logger().Debug("debug")
	ctx.Logger().Info("info")
	ctx.Logger().Error("error")
	require.Equal(t, *logger.logs, []string{"debug", "info", "error"})
}

type dummy int64

func (d dummy) Clone() interface{} {
	return d
}

// Testing saving/loading primitive values to/from the context
func TestContextWithPrimitive(t *testing.T) {
	ctx := context.NewContext(nil, abci.Header{}, false, log.NewNopLogger(), nil)

	clonerkey := "cloner"
	stringkey := "string"
	int32key := "int32"
	uint32key := "uint32"
	uint64key := "uint64"

	keys := []string{clonerkey, stringkey, int32key, uint32key, uint64key}

	for _, key := range keys {
		require.Nil(t, ctx.Value(key))
	}

	clonerval := dummy(1)
	stringval := "string"
	int32val := int32(1)
	uint32val := uint32(2)
	uint64val := uint64(3)

	ctx = ctx.
		WithCloner(clonerkey, clonerval).
		WithString(stringkey, stringval).
		WithInt32(int32key, int32val).
		WithUint32(uint32key, uint32val).
		WithUint64(uint64key, uint64val)

	require.Equal(t, clonerval, ctx.Value(clonerkey))
	require.Equal(t, stringval, ctx.Value(stringkey))
	require.Equal(t, int32val, ctx.Value(int32key))
	require.Equal(t, uint32val, ctx.Value(uint32key))
	require.Equal(t, uint64val, ctx.Value(uint64key))
}

// Testing saving/loading sdk type values to/from the context
func TestContextWithCustom(t *testing.T) {
	var ctx context.Context
	require.True(t, ctx.IsZero())

	require.Panics(t, func() { ctx.BlockHeader() })
	require.Panics(t, func() { ctx.BlockHeight() })
	require.Panics(t, func() { ctx.ChainID() })
	require.Panics(t, func() { ctx.TxBytes() })
	require.Panics(t, func() { ctx.Logger() })
	require.Panics(t, func() { ctx.SigningValidators() })
	require.Panics(t, func() { ctx.GasMeter() })

	header := abci.Header{}
	height := int64(1)
	chainid := "chainid"
	ischeck := true
	txbytes := []byte("txbytes")
	logger := NewMockLogger()
	signvals := []abci.SigningValidator{{}}
	meter := types.NewGasMeter(10000)
	minFees := make([]types.Coin, 1)
	blockTxIndex := int64(100)

	handerler := func(ctx context.Context, itx interface{}) types.Result {
		return types.Result{}
	}

	ctx = context.NewContext(nil, header, ischeck, logger, nil).
		WithBlockHeight(height).
		WithChainID(chainid).
		WithTxBytes(txbytes).
		WithSigningValidators(signvals).
		WithGasMeter(meter).
		WithMinimumFees(minFees).
		WithBlockTxIndex(blockTxIndex).
		WithTxQcpResultHandler(handerler)

	require.Equal(t, header, ctx.BlockHeader())
	require.Equal(t, height, ctx.BlockHeight())
	require.Equal(t, chainid, ctx.ChainID())
	require.Equal(t, ischeck, ctx.IsCheckTx())
	require.Equal(t, txbytes, ctx.TxBytes())
	require.Equal(t, logger, ctx.Logger())
	require.Equal(t, signvals, ctx.SigningValidators())
	require.Equal(t, meter, ctx.GasMeter())
	require.Equal(t, blockTxIndex, ctx.BlockTxIndex())

	h := ctx.TxQcpResultHandler()

	// tye := reflect.TypeOf(h)

	fmt.Println(h(ctx, ctx))

	ctx = context.NewContext(nil, header, ischeck, logger, nil)

	index := ctx.BlockTxIndex()
	require.Equal(t, int64(-1), index)

	ctx = ctx.WithBlockTxIndex(1)
	index = ctx.BlockTxIndex()
	require.Equal(t, int64(1), index)

}
