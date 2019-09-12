package qcp

import (
	"fmt"

	"github.com/QOSGroup/qbase/mapper"
	"github.com/QOSGroup/qbase/txs"
	go_amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"
)

const (
	MapperName = "qcp"
	//需要输出到"chainId"的qcp tx最大序号
	outSequencePrefixKey = "sequence/out/"
	outSequenceKey       = outSequencePrefixKey + "%s"
	//需要输出到"chainId"的每个qcp tx
	outSequenceTxPrefixKey = "tx/out/"
	outSequenceTxKey       = "outSequenceTxPrefixKey" + "%s/%d"
	//已经接受到来自"chainId"的qcp 的合法公钥tx最大序号
	inSequencePrefixKey = "sequence/in/"
	inSequenceKey       = inSequencePrefixKey + "%s"
	//接受来自"chainId"
	inPubkeyPrefixKey = "pubkey/in/"
	inPubkeyKey       = inPubkeyPrefixKey + "%s"
)

func BuildQcpStoreQueryPath() []byte {
	return []byte(fmt.Sprintf("/store/%s/key", MapperName))
}

func BuildOutSequenceKey(outChainID string) []byte {
	return []byte(fmt.Sprintf(outSequenceKey, outChainID))
}

func BuildOutSequencePrefixKey() []byte {
	return []byte(outSequencePrefixKey)
}

func BuildOutSequenceTxKey(outChainID string, sequence int64) []byte {
	return []byte(fmt.Sprintf(outSequenceTxKey, outChainID, sequence))
}

func BuildOutSequenceTxPrefixKey() []byte {
	return []byte(outSequenceTxPrefixKey)
}

func BuildInSequenceKey(inChainID string) []byte {
	return []byte(fmt.Sprintf(inSequenceKey, inChainID))
}

func BuildInSequencePrefixKey() []byte {
	return []byte(inSequencePrefixKey)
}

func BuildInPubkeyKey(inChainID string) []byte {
	return []byte(fmt.Sprintf(inPubkeyKey, inChainID))
}

func BuildInPubkeyPrefixKey() []byte {
	return []byte(inPubkeyPrefixKey)
}

type QcpMapper struct {
	*mapper.BaseMapper
}

var _ mapper.IMapper = (*QcpMapper)(nil)

func NewQcpMapper(cdc *go_amino.Codec) *QcpMapper {
	var qcpMapper = QcpMapper{}
	qcpMapper.BaseMapper = mapper.NewBaseMapper(cdc, MapperName)
	return &qcpMapper
}

func (mapper *QcpMapper) Copy() mapper.IMapper {
	cpyMapper := &QcpMapper{}
	cpyMapper.BaseMapper = mapper.BaseMapper.Copy()
	return cpyMapper
}

func (mapper *QcpMapper) GetChainInTrustPubKey(inChain string) (pubkey crypto.PubKey) {
	mapper.Get(BuildInPubkeyKey(inChain), &pubkey)
	return
}

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
