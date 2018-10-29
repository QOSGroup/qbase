package baseabci

import (
	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/context"
	"github.com/QOSGroup/qbase/qcp"
	"github.com/QOSGroup/qbase/txs"
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

//see: handler.go: TxQcpResultHandler
func ConvertTxQcpResult(txQcpResult interface{}) (*txs.QcpTxResult, bool) {
	qcpResult, ok := txQcpResult.(*txs.QcpTxResult)
	return qcpResult, ok
}
