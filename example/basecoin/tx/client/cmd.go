package client

import (
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
)

func Commands(cdc *amino.Codec) []*cobra.Command {
	return []*cobra.Command{
		sendTxCmd(cdc),
	}
}
