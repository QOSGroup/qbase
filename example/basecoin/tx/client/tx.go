package client

import (
	"github.com/QOSGroup/qbase/client/account"
	"github.com/QOSGroup/qbase/client/context"
	btx "github.com/QOSGroup/qbase/client/tx"
	"github.com/QOSGroup/qbase/example/basecoin/tx"
	"github.com/QOSGroup/qbase/txs"
	btypes "github.com/QOSGroup/qbase/types"
	"github.com/tendermint/go-amino"

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

			return btx.BroadcastTxAndPrintResult(cdc, func(ctx context.CLIContext) (txs.ITx, error) {
				fromAddr, err := account.GetAddrFromFlag(ctx, flagFrom)
				if err != nil {
					return nil, err
				}
				toAddr, err := account.GetAddrFromFlag(ctx, flagTo)
				if err != nil {
					return nil, err
				}

				name := viper.GetString(flagCoinName)
				amount := viper.GetInt64(flagCoinAmount)
				if err != nil {
					return nil, err
				}
				sendTx := tx.NewSendTx(fromAddr, toAddr, btypes.Coin{name, btypes.NewInt(amount)})
				return &sendTx, nil

			})
		},
	}

	cmd.Flags().String(flagFrom, "", "Address to send coins")
	cmd.Flags().String(flagTo, "", "Address to receive coins")
	cmd.Flags().String(flagCoinName, "", "Name of coin to send")
	cmd.Flags().Int64(flagCoinAmount, 0, "Amount of coin to send")

	cmd.MarkFlagRequired(flagFrom)
	cmd.MarkFlagRequired(flagTo)

	return cmd
}
