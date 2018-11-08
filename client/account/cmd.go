package account

import (
	"fmt"
	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/types"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
)

func QueryAccountCmd(cdc *amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account [addr]",
		Short: "query account by address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			addr, _ := types.GetAddrFromBech32(args[0])
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			output, err := queryAccount(cliCtx, addr)
			if err != nil {
				return err
			}
			fmt.Println(cliCtx.ToJSONIndentStr(output))
			return nil
		},
	}

	return cmd
}
