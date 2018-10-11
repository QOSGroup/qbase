package qcp

import (
	"fmt"
	"testing"

	"github.com/QOSGroup/qbase/mapper"

	"github.com/QOSGroup/qbase/context"
	"github.com/QOSGroup/qbase/store"
	"github.com/QOSGroup/qbase/txs"
	"github.com/stretchr/testify/require"
	go_amino "github.com/tendermint/go-amino"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/encoding/amino"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

func Test_qcpMapper_GetMaxChainOutSequence(t *testing.T) {

	cdc := defaultCdc()
	qcpMapper := NewQcpMapper()

	qcpMapper.SetCodec(cdc)

	storeKey := qcpMapper.GetStoreKey()
	ctx := defaultContext(storeKey)

	mapper := make(map[string]mapper.IMapper)
	mapper[qcpMapper.Name()] = qcpMapper

	ctx = ctx.WithRegisteredMap(mapper)

	qcpMapper, _ = ctx.Mapper(qcpMapper.Name()).(*QcpMapper)

	outChain := "qsc"
	seq := int64(12)
	cseq := int64(13)

	maxSeq := qcpMapper.GetMaxChainOutSequence(outChain)
	require.Equal(t, int64(0), maxSeq)

	qcpMapper.SetMaxChainOutSequence(outChain, seq)
	maxSeq = qcpMapper.GetMaxChainOutSequence(outChain)
	require.Equal(t, seq, maxSeq)

	//new cache context
	cctx, write := ctx.CacheContext()

	fmt.Println(ctx.KVStore(storeKey))
	fmt.Println(cctx.KVStore(storeKey))

	newQcpMapper, _ := cctx.Mapper(qcpMapper.Name()).(*QcpMapper)

	fmt.Println(qcpMapper)
	fmt.Println(newQcpMapper)

	maxSeq = newQcpMapper.GetMaxChainOutSequence(outChain)
	require.Equal(t, seq, maxSeq)

	var seq1 int64
	store := cctx.KVStore(storeKey)
	c := store.Get([]byte(fmt.Sprintf(outSequenceKey, outChain)))
	qcpMapper.GetCodec().UnmarshalBinaryBare(c, &seq1)

	store = ctx.KVStore(storeKey)
	c = store.Get([]byte(fmt.Sprintf(outSequenceKey, outChain)))
	qcpMapper.GetCodec().UnmarshalBinaryBare(c, &seq1)

	newQcpMapper.SetMaxChainOutSequence(outChain, cseq)
	maxSeq = newQcpMapper.GetMaxChainOutSequence(outChain)
	require.Equal(t, cseq, maxSeq)

	store = cctx.KVStore(storeKey)
	c = store.Get([]byte(fmt.Sprintf(outSequenceKey, outChain)))
	qcpMapper.GetCodec().UnmarshalBinaryBare(c, &seq1)

	store = ctx.KVStore(storeKey)
	c = store.Get([]byte(fmt.Sprintf(outSequenceKey, outChain)))
	qcpMapper.GetCodec().UnmarshalBinaryBare(c, &seq1)

	//重置qcpMapper中的kestore
	// qcpMapper.SetStore(ctx.KVStore(storeKey))
	maxSeq = qcpMapper.GetMaxChainOutSequence(outChain)
	require.Equal(t, seq, maxSeq)

	write()

	maxSeq = qcpMapper.GetMaxChainOutSequence(outChain)
	require.Equal(t, cseq, maxSeq)

}

func defaultCdc() *go_amino.Codec {
	var cdc = go_amino.NewCodec()
	cryptoAmino.RegisterAmino(cdc)
	cdc.RegisterConcrete(&txs.QcpTxResult{}, "qbase/txs/qcpresult", nil)
	cdc.RegisterConcrete(&txs.Signature{}, "qbase/txs/qcptx", nil)
	cdc.RegisterConcrete(&txs.TxStd{}, "qbase/txs/stdtx", nil)
	cdc.RegisterInterface((*txs.ITx)(nil), nil)
	return cdc
}

func defaultContext(key store.StoreKey) context.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, store.StoreTypeIAVL, db)
	cms.LoadLatestVersion()

	ctx := context.NewContext(cms, abci.Header{}, false, log.NewNopLogger(), nil)
	return ctx
}
