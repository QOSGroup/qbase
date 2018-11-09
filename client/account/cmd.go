package account

import (
	"fmt"
	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/client/keys"
	"github.com/QOSGroup/qbase/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/go-amino"
)

const (
	flagName    = "name"
	flagAddress = "addr"
)

func QueryAccountCmd(cdc *amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "query account by address or name",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			addrStr := viper.GetString(flagAddress)
			var addr types.Address
			if len(addrStr) != 0 {
				address, err := types.GetAddrFromBech32(args[0])
				if err != nil {
					return err
				}
				addr = address
			} else {
				name := viper.GetString(flagName)
				info, err := keys.GetKeyInfo(cliCtx, name)
				if err != nil {
					return nil
				}
				addr = info.GetAddress()
			}

			output, err := queryAccount(cliCtx, addr)
			if err != nil {
				return err
			}
			fmt.Println(cliCtx.ToJSONIndentStr(output))
			return nil
		},
	}

	cmd.Flags().String(flagName, "", "name of account")
	cmd.Flags().String(flagAddress, "", "address of account")

	return cmd
}
