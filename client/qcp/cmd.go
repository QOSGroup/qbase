package qcp

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/QOSGroup/qbase/client"
	"github.com/QOSGroup/qbase/client/context"
	go_amino "github.com/tendermint/go-amino"
)

const (
	flagOutSeq = "seq"
)

var qcpCommand = &cobra.Command{
	Use:   "qcp",
	Short: "qcp subcommands",
}

func AddCommands(cmd *cobra.Command, cdc *go_amino.Codec) {
	qcpCommand.AddCommand(
		client.GetCommands(
			QueryOutChainSeqCmd(cdc),
			QueryOutChainTxCmd(cdc),
			QueryInChainSeqCmd(cdc),
		)...,
	)
	cmd.AddCommand(
		qcpCommand,
	)
}

func QueryOutChainSeqCmd(cdc *go_amino.Codec) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "outseq",
		Short: "Get max sequence  to outChain",
		Long: strings.TrimSpace(`
Get Max Sequence  to OutChain

example:
$ basecli qcp outseq --chain-id  [outChainID]
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			outChainID := viper.GetString(client.FlagChainID)
			if outChainID == "" {
				return errors.New("missing outChainID. use `--chain-id outChainID` set params. ")
			}

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			seq, err := GetOutChainSequence(cliCtx, outChainID)

			if err != nil {
				return err
			}

			fmt.Println(seq)
			return nil
		},
	}
	return cmd
}

func QueryOutChainTxCmd(cdc *go_amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "outtx",
		Short: "Query qcp out tx ",
		Long: strings.TrimSpace(`
query qcp out tx from chainID and sequence

example:
$ basecli qcp outtx --chain-id  [outChainID] --seq [Seq]
`),
		RunE: func(cmd *cobra.Command, args []string) error {

			outChainID := viper.GetString(client.FlagChainID)
			if outChainID == "" {
				return errors.New("missing outChainID. use `--chain-id outChainID` set params. ")
			}

			seq := viper.GetInt64(flagOutSeq)

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			result, err := GetGetOutChainTx(cliCtx, outChainID, seq)

			if err != nil {
				return err
			}

			fmt.Println(cliCtx.ToJSONIndentStr(result))
			return nil
		},
	}

	cmd.Flags().Int64(flagOutSeq, 1, "out sequence")

	return cmd
}

func QueryInChainSeqCmd(cdc *go_amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inseq",
		Short: "Get max sequence received from inChain",
		Long: strings.TrimSpace(`
Get max sequence received from inChain

example:
$ basecli qcp inseq --chain-id  [inChainID]
`),
		RunE: func(cmd *cobra.Command, args []string) error {

			inChainID := viper.GetString(client.FlagChainID)
			if inChainID == "" {
				return errors.New("missing inChainID. use `--chain-id inChainID` set params. ")
			}

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			seq, err := GetInChainSequence(cliCtx, inChainID)

			if err != nil {
				return err
			}

			fmt.Println(seq)
			return nil
		},
	}

	return cmd
}
