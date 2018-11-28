package qcp

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/QOSGroup/qbase/client/context"
	go_amino "github.com/tendermint/go-amino"
)

const (
	flagOutSeq = "seq"
)

func QcpCommands(cdc *go_amino.Codec) []*cobra.Command {
	return []*cobra.Command{
		listCommand(cdc),
		outSeqCmd(cdc),
		inSeqCmd(cdc),
		outTxCmd(cdc),
	}
}

func listCommand(cdc *go_amino.Codec) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all crossQcp chain's sequence info",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			result, err := QueryQcpChainsInfo(cliCtx)
			if err != nil {
				return err
			}

			printTab(result)
			return nil
		},
	}

	return cmd
}

func printTab(res []qcpChainsResult) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "|Chain\tType\tMaxSequence\t")
	fmt.Fprintln(w, "|-----\t----\t-----------\t")
	for _, qcpRes := range res {
		fmt.Fprintf(w, "|%s\t%s\t%d\t\n", qcpRes.ChainID, qcpRes.T, qcpRes.Sequence)
	}
	w.Flush()
}

func outSeqCmd(cdc *go_amino.Codec) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "out [chainID]",
		Args:  cobra.ExactArgs(1),
		Short: "Get max sequence to outChain",
		Long: strings.TrimSpace(`
Get Max Sequence  to OutChain

example:
$ basecli qcp out [outChainID]
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			outChainID := args[0]
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

func outTxCmd(cdc *go_amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx [chainID]",
		Args:  cobra.ExactArgs(1),
		Short: "Query qcp out tx ",
		Long: strings.TrimSpace(`
query qcp out tx from chainID and sequence

example:
$ basecli qcp tx [chainID] --seq [Seq]
`),
		RunE: func(cmd *cobra.Command, args []string) error {

			outChainID := args[0]
			seq := viper.GetInt64(flagOutSeq)

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			result, err := GetGetOutChainTx(cliCtx, outChainID, seq)

			if err != nil {
				return err
			}

			return cliCtx.PrintResult(result)
		},
	}

	cmd.Flags().Int64(flagOutSeq, 1, "out sequence")
	cmd.MarkFlagRequired(flagOutSeq)
	return cmd
}

func inSeqCmd(cdc *go_amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "in [chainID]",
		Args:  cobra.ExactArgs(1),
		Short: "Get max sequence received from inChain",
		Long: strings.TrimSpace(`
Get max sequence received from inChain

example:
$ basecli qcp in  [inChainID]
`),
		RunE: func(cmd *cobra.Command, args []string) error {

			inChainID := args[0]
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
