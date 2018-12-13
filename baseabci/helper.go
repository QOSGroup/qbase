package baseabci

import (
	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/consensus"
	"github.com/QOSGroup/qbase/context"
	"github.com/QOSGroup/qbase/qcp"
	"github.com/QOSGroup/qbase/txs"
	abci "github.com/tendermint/tendermint/abci/types"
)

func GetAccountMapper(ctx context.Context) *account.AccountMapper {
	mapper := ctx.Mapper(account.AccountMapperName)
	if mapper == nil {
		return nil
	}
	return mapper.(*account.AccountMapper)
}

func GetQcpMapper(ctx context.Context) *qcp.QcpMapper {
	mapper := ctx.Mapper(qcp.QcpMapperName)
	if mapper == nil {
		return nil
	}
	return mapper.(*qcp.QcpMapper)
}

func GetConsMapper(ctx context.Context) *consensus.ConsensusMapper {
	mapper := ctx.Mapper(consensus.ConsensusMapperName)
	if mapper == nil {
		return nil
	}
	return mapper.(*consensus.ConsensusMapper)
}

//see: handler.go: TxQcpResultHandler
func ConvertTxQcpResult(txQcpResult interface{}) (*txs.QcpTxResult, bool) {
	qcpResult, ok := txQcpResult.(*txs.QcpTxResult)
	return qcpResult, ok
}

func GetConsParams(ctx context.Context) *abci.ConsensusParams {
	consMapper := GetConsMapper(ctx)
	if consMapper != nil {
		var consParam abci.ConsensusParams
		if ok := consMapper.Get(consensus.BuildConsKey(), &consParam); ok {
			return &consParam
		}
	}
	return nil
}
