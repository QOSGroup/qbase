package block

import (
	"fmt"

	"github.com/QOSGroup/qbase/client"
	"github.com/QOSGroup/qbase/client/context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	go_amino "github.com/tendermint/go-amino"
)

func statusCommand(cdc *go_amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Query remote node for status",
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.Set(client.FlagTrustNode, true)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			node, err := cliCtx.GetNode()
			if err != nil {
				return err
			}

			status, err := node.Status()
			if err != nil {
				return err
			}

			output, err := cliCtx.ToJSONIndentStr(status)
			fmt.Println(string(output))
			return nil
		},
	}

	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	viper.BindPFlag(client.FlagNode, cmd.Flags().Lookup(client.FlagNode))
	return cmd
}
