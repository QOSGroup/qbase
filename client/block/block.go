package block

import (
	"fmt"
	"strconv"

	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/client/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	go_amino "github.com/tendermint/go-amino"
)

func blockCommand(cdc *go_amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block [height]",
		Short: "Get block info at given height",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			h, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			height := int64(h)

			viper.Set(types.FlagTrustNode, true)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			node, err := cliCtx.GetNode()
			if err != nil {
				return err
			}

			res, err := node.Block(&height)
			if err != nil {
				return err
			}

			output, err := cliCtx.ToJSONIndentStr(res)
			fmt.Println(string(output))

			return nil
		},
	}

	cmd.Flags().StringP(types.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	viper.BindPFlag(types.FlagNode, cmd.Flags().Lookup(types.FlagNode))
	return cmd
}
