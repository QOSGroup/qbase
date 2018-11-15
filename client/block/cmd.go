package block

import (
	"github.com/QOSGroup/qbase/client/types"
	"github.com/spf13/cobra"
	go_amino "github.com/tendermint/go-amino"
)

func InternalBlockCommand(cdc *go_amino.Codec) []*cobra.Command {

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
