package qcp

import (
	"fmt"

	ctx "github.com/QOSGroup/qbase/context"
	"github.com/QOSGroup/qbase/mapper"
	"github.com/QOSGroup/qbase/store"
	"github.com/QOSGroup/qbase/txs"
	"github.com/tendermint/tendermint/crypto"
)

const (
	QcpMapperName    = "qcpmapper"
	storeKey         = "qcp"
	outSequenceKey   = "[%s]/out/sequence"
	outSequenceTxKey = "[%s]/out/tx_[%d]"
	inSequenceKey    = "[%s]/in/sequence"
	inPubkeyKey      = "[%s]/in/pubkey"
)

type QcpMapper struct {
	*mapper.BaseMapper
}

var _ mapper.IMapper = (*QcpMapper)(nil)

func NewQcpMapper() *QcpMapper {
	var qcpMapper = QcpMapper{}
	qcpMapper.BaseMapper = mapper.NewBaseMapper(store.NewKVStoreKey(storeKey))
	return &qcpMapper
}

func (mapper *QcpMapper) Copy() mapper.IMapper {
	cpyMapper := &QcpMapper{}
	cpyMapper.BaseMapper = mapper.BaseMapper.Copy()
	return cpyMapper
}

func (mapper *QcpMapper) Name() string {
	return QcpMapperName
}

//TODO: test case
func (mapper *QcpMapper) GetChainInTruestPubKey(inChain string) (pubkey crypto.PubKey) {
	key := fmt.Sprintf(inPubkeyKey, inChain)
	mapper.Get([]byte(key), &pubkey)
	return
}

//TODO: test case
func (mapper *QcpMapper) SetChainInTruestPubKey(inChain string, pubkey crypto.PubKey) {
	key := fmt.Sprintf(inPubkeyKey, inChain)
	mapper.Set([]byte(key), pubkey)
}

func (mapper *QcpMapper) GetMaxChainOutSequence(outChain string) (seq int64) {
	key := fmt.Sprintf(outSequenceKey, outChain)
	mapper.Get([]byte(key), &seq)
	return
}

func (mapper *QcpMapper) SetMaxChainOutSequence(outChain string, sequence int64) {
	key := fmt.Sprintf(outSequenceKey, outChain)
	mapper.Set([]byte(key), sequence)
}

func (mapper *QcpMapper) GetChainOutTxs(outChain string, sequence int64) (txQcp *txs.TxQcp) {
	key := fmt.Sprintf(outSequenceTxKey, outChain, sequence)
	mapper.Get([]byte(key), txQcp)
	return
}

func (mapper *QcpMapper) SetChainOutTxs(outChain string, sequence int64, txQcp *txs.TxQcp) {
	key := fmt.Sprintf(outSequenceTxKey, outChain, sequence)
	mapper.Set([]byte(key), *txQcp)
}

func (mapper *QcpMapper) GetMaxChainInSequence(inChain string) (seq int64) {
	key := fmt.Sprintf(inSequenceKey, inChain)
	mapper.Get([]byte(key), &seq)
	return
}

func (mapper *QcpMapper) SetMaxChainInSequence(inChain string, sequence int64) {
	key := fmt.Sprintf(inSequenceKey, inChain)
	mapper.Set([]byte(key), sequence)
}

//TODO: test case
func (mapper *QcpMapper) SaveCrossChainResult(ctx ctx.Context, payload txs.TxStd, toChainID string, isResult bool, signer crypto.PrivKey) (txQcp *txs.TxQcp) {

	maxSequence := mapper.GetMaxChainOutSequence(toChainID)

	txQcp = &txs.TxQcp{
		Payload:     payload,
		From:        ctx.ChainID(),
		To:          toChainID,
		Sequence:    maxSequence + 1,
		Sig:         txs.Signature{},
		BlockHeight: ctx.BlockHeight(),
		TxIndx:      ctx.BlockTxIndex(),
		IsResult:    isResult,
	}

	if signer != nil {
		signature, err := signer.Sign(txQcp.GetSigData())
		if err != nil {
			panic("sign txQcp error")
		}
		txQcp.Sig.Signature = signature
		txQcp.Sig.Pubkey = signer.PubKey()
	}

	mapper.SetMaxChainOutSequence(toChainID, maxSequence+1)
	mapper.SetChainOutTxs(toChainID, maxSequence+1, txQcp)

	return
}
