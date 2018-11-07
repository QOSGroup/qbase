package qcp

import (
	"fmt"

	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/qcp"
	"github.com/QOSGroup/qbase/txs"
)

func query(ctx context.CLIContext, key []byte) ([]byte, error) {
	path := qcp.BuildQcpStoreQueryPath()
	return ctx.Query(string(path), key)
}

func GetOutChainSequence(ctx context.CLIContext, outChainID string) (int64, error) {
	key := qcp.BuildOutSequenceKey(outChainID)
	bz, err := query(ctx, key)

	if err != nil {
		return 0, err
	}

	if len(bz) == 0 {
		return 0, fmt.Errorf("GetOutChainSequence return empty. there is not exists %s out sequence", outChainID)
	}

	var seq int64
	err = ctx.Codec.UnmarshalBinaryBare(bz, &seq)
	if err != nil {
		return 0, err
	}

	return seq, nil
}

func GetGetOutChainTx(ctx context.CLIContext, outChainID string, seq int64) (*txs.TxQcp, error) {
	key := qcp.BuildOutSequenceTxKey(outChainID, seq)
	bz, err := query(ctx, key)

	if err != nil {
		return nil, err
	}

	if len(bz) == 0 {
		return nil, fmt.Errorf("GetGetOutChainTx return empty. there is not exists %s/%d out tx", outChainID, seq)
	}

	var tx txs.TxQcp
	err = ctx.Codec.UnmarshalBinaryBare(bz, &tx)
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

func GetInChainSequence(ctx context.CLIContext, inChainID string) (int64, error) {
	key := qcp.BuildInSequenceKey(inChainID)
	bz, err := query(ctx, key)

	if err != nil {
		return 0, err
	}

	if len(bz) == 0 {
		return 0, fmt.Errorf("GetInChainSequence return empty. there is not exists %s in sequence", key)
	}

	var seq int64
	err = ctx.Codec.UnmarshalBinaryBare(bz, &seq)
	if err != nil {
		return 0, err
	}

	return seq, nil
}
