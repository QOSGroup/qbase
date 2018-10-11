package types

import (
	go_amino "github.com/tendermint/go-amino"
)

//Tx: 对stdTx及qcpTx类型的封装
type Tx interface {
	Type() string
}

// TxDecoder unmarshals transaction bytes
type TxDecoder func(txBytes []byte) (Tx, Error)

func GetTxDecoder(cdc *go_amino.Codec) TxDecoder {
	return func(txBytes []byte) (Tx, Error) {
		var tx Tx
		err := cdc.UnmarshalBinaryBare(txBytes, &tx)

		if err != nil {
			return nil, ErrInternal("txBytes decoder failed")
		}

		return tx, nil
	}
}
