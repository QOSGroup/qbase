package account

import (
	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/types"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
)

func QueryAccountCmd(cdc *amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account [name or address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query account info by address or name",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			var addr types.Address
			addr, err := GetAddrFromValue(cliCtx, args[0])
			if err != nil {
				return err
			}

			output, err := queryAccount(cliCtx, addr)
			if err != nil {
				return err
			}

			return cliCtx.PrintResult(output)
		},
	}

	return cmd
}
