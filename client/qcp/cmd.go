package qcp

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/QOSGroup/qbase/client/context"
	go_amino "github.com/tendermint/go-amino"
)

const (
	flagOutSeq = "seq"
)

func QueryOutChainSeqCmd(cdc *go_amino.Codec) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "outseq [chainID]",
		Args:  cobra.ExactArgs(1),
		Short: "Get max sequence to outChain",
		Long: strings.TrimSpace(`
Get Max Sequence  to OutChain

example:
$ basecli qcp outseq [outChainID]
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

func QueryOutChainTxCmd(cdc *go_amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "outtx [chainID]",
		Args:  cobra.ExactArgs(1),
		Short: "Query qcp out tx ",
		Long: strings.TrimSpace(`
query qcp out tx from chainID and sequence

example:
$ basecli qcp outtx [chainID] --seq [Seq]
`),
		RunE: func(cmd *cobra.Command, args []string) error {

			outChainID := args[0]
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
	cmd.MarkFlagRequired(flagOutSeq)
	return cmd
}

func QueryInChainSeqCmd(cdc *go_amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inseq [chainID]",
		Args:  cobra.ExactArgs(1),
		Short: "Get max sequence received from inChain",
		Long: strings.TrimSpace(`
Get max sequence received from inChain

example:
$ basecli qcp inseq  [inChainID]
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
