package block

import (
	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/client/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	go_amino "github.com/tendermint/go-amino"
)

func statusCommand(cdc *go_amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Query remote node for status",
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.Set(types.FlagTrustNode, true)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			node, err := cliCtx.GetNode()
			if err != nil {
				return err
			}

			status, err := node.Status()
			if err != nil {
				return err
			}

			return cliCtx.PrintResult(status)
		},
	}

	cmd.Flags().StringP(types.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	cmd.Flags().Bool(types.FlagJSONIndet, false, "print indent result json")
	viper.BindPFlag(types.FlagNode, cmd.Flags().Lookup(types.FlagNode))

	return cmd
}
