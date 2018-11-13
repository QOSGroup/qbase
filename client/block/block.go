package block

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/QOSGroup/qbase/client"
	"github.com/QOSGroup/qbase/client/context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	go_amino "github.com/tendermint/go-amino"
)

func blockCommand(cdc *go_amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block [height]",
		Short: "Get block info at given height",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) < 1 {
				return errors.New("missing height args")
			}

			h, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			height := int64(h)

			viper.Set(client.FlagTrustNode, true)
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

	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	viper.BindPFlag(client.FlagNode, cmd.Flags().Lookup(client.FlagNode))
	return cmd
}
