package tx

import (
	"fmt"
	"github.com/QOSGroup/qbase/client/context"
	cliTypes "github.com/QOSGroup/qbase/client/types"
	"github.com/QOSGroup/qbase/types"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
	"io/ioutil"
)

func BroadcastCmd(cdc *amino.Codec) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "broadcast [file]",
		Short: "broadcast signed file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)
			txBytes, err := ioutil.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read signed file err. err: %s", err.Error())
			}

			var tx types.Tx
			err = cdc.UnmarshalJSON(txBytes, &tx)
			if err != nil {
				return fmt.Errorf("signed file UnmarshalJSON err. err: %s", err.Error())
			}

			txAminoBytes, err := cdc.MarshalBinaryBare(tx)
			if err != nil {
				return fmt.Errorf("signed file MarshalBinaryBare err. err: %s", err.Error())
			}

			result, err := ctx.BroadcastTx(txAminoBytes)
			ctx.PrintResult(result)

			return err
		},
	}

	return cliTypes.PostCommands(cmd)[0]
}
