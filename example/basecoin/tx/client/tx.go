package client

import (
	"fmt"
	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/client/keys"
	btx "github.com/QOSGroup/qbase/client/tx"
	"github.com/QOSGroup/qbase/example/basecoin/tx"
	"github.com/QOSGroup/qbase/txs"
	btypes "github.com/QOSGroup/qbase/types"
	"github.com/tendermint/go-amino"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagFrom       = "from"
	flagTo         = "to"
	flagCoinName   = "coin-name"
	flagCoinAmount = "coin-amount"

	flagQCPChain = "qcp-chain"
)

// 链内交易
func sendTxCmd(cdc *amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Create and sign a send tx",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			result, err := btx.BroadcastSignedTx(cliCtx, func(ctx context.CLIContext) (txs.ITx, error) {
				fromName := viper.GetString(flagFrom)
				fromInfo, err := keys.GetKeyInfo(cliCtx, fromName)
				if err != nil {
					return nil, err
				}

				toName := viper.GetString(flagTo)
				toInfo, err := keys.GetKeyInfo(cliCtx, toName)
				if err != nil {
					return nil, err
				}

				name := viper.GetString(flagCoinName)
				amount, err := strconv.ParseInt(viper.GetString(flagCoinAmount), 10, 64)
				if err != nil {
					return nil, err
				}
				sendTx := tx.NewSendTx(fromInfo.GetAddress(), toInfo.GetAddress(), btypes.BaseCoin{name, btypes.NewInt(amount)})
				return &sendTx, nil
			})

			msg, _ := cdc.MarshalJSON(result)
			fmt.Println(string(msg))

			return err
		},
	}

	cmd.Flags().String(flagFrom, "", "Address to send coins")
	cmd.Flags().String(flagTo, "", "Address to receive coins")
	cmd.Flags().String(flagCoinName, "", "Name of coin to send")
	cmd.Flags().String(flagCoinAmount, "", "Amount of coin to send")

	cmd.MarkFlagRequired(flagFrom)
	cmd.MarkFlagRequired(flagTo)

	return cmd
}
