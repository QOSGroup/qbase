package block

import (
	"github.com/QOSGroup/qbase/client/types"
	"github.com/spf13/cobra"
	go_amino "github.com/tendermint/go-amino"
)

func QueryCommand(cdc *go_amino.Codec) []*cobra.Command {
	return []*cobra.Command{
		storeCommand(cdc),
		consensusCommand(cdc),
	}
}

func BlockCommand(cdc *go_amino.Codec) []*cobra.Command {

	return []*cobra.Command{
		statusCommand(cdc),
		types.LineBreak,
		validatorsCommand(cdc),
		blockCommand(cdc),
		types.LineBreak,
		searchTxCmd(cdc),
		queryTxCmd(cdc),
	}
}
