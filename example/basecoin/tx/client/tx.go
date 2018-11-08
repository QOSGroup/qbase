package client

import (
	"fmt"
	cliacc "github.com/QOSGroup/qbase/client/account"
	"github.com/QOSGroup/qbase/client/context"
	cliqcp "github.com/QOSGroup/qbase/client/qcp"
	btypes "github.com/QOSGroup/qbase/types"
	"github.com/tendermint/go-amino"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagFrom       = "from"
	flagPriKey     = "from-prikey"
	flagTo         = "to"
	flagCoinName   = "coin-name"
	flagCoinAmount = "coin-amount"

	flagQCPChain  = "qcp-chain"
	flagQCPPriKey = "qcp-prikey"
)

// 链内交易
func SendTxCmd(cdc *amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Create and sign a send tx",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().
				WithCodec(cdc)

			fromStr := viper.GetString(flagFrom)
			fromAddr, err := btypes.GetAddrFromBech32(fromStr)
			if err != nil {
				return err
			}
			from, err := cliacc.GetAccount(cliCtx, fromAddr)
			if err != nil {
				return err
			}

			fromPri := viper.GetString(flagPriKey)

			toStr := viper.GetString(flagTo)
			toAddr, err := btypes.GetAddrFromBech32(toStr)
			if err != nil {
				return err
			}

			name := viper.GetString(flagCoinName)
			amount, err := strconv.ParseInt(viper.GetString(flagCoinAmount), 10, 64)

			tx := BuildStdTx(cdc, from, fromPri, toAddr, btypes.BaseCoin{name, btypes.NewInt(amount)})

			result, err := cliCtx.BroadcastTx(cdc.MustMarshalBinaryBare(tx))

			msg, _ := cdc.MarshalJSON(result)
			fmt.Println(string(msg))

			return err
		},
	}

	cmd.Flags().String(flagFrom, "", "Address to send coins")
	cmd.Flags().String(flagPriKey, "", "Sender's PriKey")
	cmd.Flags().String(flagTo, "", "Address to receive coins")
	cmd.Flags().String(flagCoinName, "", "Name of coin to send")
	cmd.Flags().String(flagCoinAmount, "", "Amount of coin to send")

	return cmd
}

// 跨链交易
func SendQCPTxCmd(cdc *amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-qcp",
		Short: "Create and sign a qcp send tx",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().
				WithCodec(cdc)

			fromStr := viper.GetString(flagFrom)
			fromAddr, err := btypes.GetAddrFromBech32(fromStr)
			if err != nil {
				return err
			}
			from, err := cliacc.GetAccount(cliCtx, fromAddr)
			if err != nil {
				return err
			}

			fromPri := viper.GetString(flagPriKey)

			toStr := viper.GetString(flagTo)
			toAddr, err := btypes.GetAddrFromBech32(toStr)
			if err != nil {
				return err
			}

			name := viper.GetString(flagCoinName)
			amount, err := strconv.ParseInt(viper.GetString(flagCoinAmount), 10, 64)

			chainId := viper.GetString(flagQCPChain)
			seq, _ := cliqcp.GetInChainSequence(cliCtx, chainId)

			chainPri := viper.GetString(flagQCPPriKey)

			stdTx := BuildStdTx(cdc, from, fromPri, toAddr, btypes.BaseCoin{name, btypes.NewInt(amount)})
			qcpTx := BuildQCPTx(cdc, stdTx, chainId, chainPri, seq + 1)

			result, err := cliCtx.BroadcastTx(cdc.MustMarshalBinaryBare(qcpTx))

			msg, _ := cdc.MarshalJSON(result)
			fmt.Println(string(msg))

			return err
		},
	}

	cmd.Flags().String(flagFrom, "", "Address to send coins")
	cmd.Flags().String(flagPriKey, "", "Sender's PriKey")
	cmd.Flags().String(flagTo, "", "Address to receive coins")
	cmd.Flags().String(flagCoinName, "", "Name of coin to send")
	cmd.Flags().String(flagCoinAmount, "", "Amount of coin to send")
	cmd.Flags().String(flagQCPChain, "", "qcp chain id")
	cmd.Flags().String(flagQCPPriKey, "", "qcp chain prikey")

	return cmd
}
