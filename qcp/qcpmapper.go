package qcp

import (
	"fmt"

	"github.com/QOSGroup/qbase/mapper"
	"github.com/QOSGroup/qbase/txs"
	go_amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"
)

const (
	QcpMapperName = "qcp"
	//需要输出到"chainId"的qcp tx最大序号
	outSequenceKey = "sequence/out/%s"
	//需要输出到"chainId"的每个qcp tx
	outSequenceTxKey = "tx/out/%s/%d"
	//已经接受到来自"chainId"的qcp 的合法公钥tx最大序号
	inSequenceKey = "sequence/in/%s"
	//接受来自"chainId"
	inPubkeyKey = "pubkey/in/%s"
)

func BuildQcpStoreQueryPath() []byte {
	return []byte(fmt.Sprintf("/store/%s/key", QcpMapperName))
}

func BuildOutSequenceKey(outChainID string) []byte {
	return []byte(fmt.Sprintf(outSequenceKey, outChainID))
}

func BuildOutSequenceTxKey(outChainID string, sequence int64) []byte {
	return []byte(fmt.Sprintf(outSequenceTxKey, outChainID, sequence))
}

func BuildInSequenceKey(inChainID string) []byte {
	return []byte(fmt.Sprintf(inSequenceKey, inChainID))
}

func BuildInPubkeyKey(inChainID string) []byte {
	return []byte(fmt.Sprintf(inPubkeyKey, inChainID))
}

type QcpMapper struct {
	*mapper.BaseMapper
}

var _ mapper.IMapper = (*QcpMapper)(nil)

func NewQcpMapper(cdc *go_amino.Codec) *QcpMapper {
	var qcpMapper = QcpMapper{}
	qcpMapper.BaseMapper = mapper.NewBaseMapper(cdc, QcpMapperName)
	return &qcpMapper
}

func (mapper *QcpMapper) Copy() mapper.IMapper {
	cpyMapper := &QcpMapper{}
	cpyMapper.BaseMapper = mapper.BaseMapper.Copy()
	return cpyMapper
}

//TODO: test case
func (mapper *QcpMapper) GetChainInTrustPubKey(inChain string) (pubkey crypto.PubKey) {
	mapper.Get(BuildInPubkeyKey(inChain), &pubkey)
	return
}

//TODO: test case
func (mapper *QcpMapper) SetChainInTrustPubKey(inChain string, pubkey crypto.PubKey) {
	mapper.Set(BuildInPubkeyKey(inChain), pubkey)
}

func (mapper *QcpMapper) GetMaxChainOutSequence(outChain string) (seq int64) {
	mapper.Get(BuildOutSequenceKey(outChain), &seq)
	return
}

func (mapper *QcpMapper) SetMaxChainOutSequence(outChain string, sequence int64) {
	mapper.Set(BuildOutSequenceKey(outChain), sequence)
}

func (mapper *QcpMapper) GetChainOutTxs(outChain string, sequence int64) *txs.TxQcp {
	var txQcp txs.TxQcp
	mapper.Get(BuildOutSequenceTxKey(outChain, sequence), &txQcp)
	return &txQcp
}

func (mapper *QcpMapper) SetChainOutTxs(outChain string, sequence int64, txQcp *txs.TxQcp) {
	mapper.Set(BuildOutSequenceTxKey(outChain, sequence), *txQcp)
}

func (mapper *QcpMapper) GetMaxChainInSequence(inChain string) (seq int64) {
	mapper.Get(BuildInSequenceKey(inChain), &seq)
	return
}

func (mapper *QcpMapper) SetMaxChainInSequence(inChain string, sequence int64) {
	mapper.Set(BuildInSequenceKey(inChain), sequence)
}

//TODO: test case
func (mapper *QcpMapper) SignAndSaveTxQcp(txQcp *txs.TxQcp, signer crypto.PrivKey) *txs.TxQcp {

	toChainID := txQcp.To
	maxSequence := mapper.GetMaxChainOutSequence(toChainID)
	txQcp.Sequence = maxSequence + 1

	if signer != nil {
		signature, err := txQcp.SignTx(signer)
		if err != nil {
			panic("sign txQcp error")
		}
		txQcp.Sig.Signature = signature
		txQcp.Sig.Pubkey = signer.PubKey()
	}

	mapper.SetMaxChainOutSequence(toChainID, maxSequence+1)
	mapper.SetChainOutTxs(toChainID, maxSequence+1, txQcp)

	return txQcp
}
