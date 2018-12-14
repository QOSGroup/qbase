package block

import (
	"errors"

	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/client/types"
	"github.com/QOSGroup/qbase/consensus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	go_amino "github.com/tendermint/go-amino"
	abci "github.com/tendermint/tendermint/abci/types"
)

func consensusCommand(cdc *go_amino.Codec) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "consensus",
		Short: "Query consensus params",
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.Set(types.FlagTrustNode, true)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			node, err := cliCtx.GetNode()
			if err != nil {
				return err
			}

			result, err := node.ABCIQuery(string(consensus.BuildConsStoreQueryPath()), consensus.BuildConsKey())
			if err != nil {
				return err
			}

			valueBz := result.Response.GetValue()
			if len(valueBz) == 0 {
				return errors.New("response empty value")
			}

			var consParams abci.ConsensusParams
			err = cdc.UnmarshalBinaryBare(valueBz, &consParams)
			if err != nil {
				return err
			}

			return cliCtx.PrintResult(consParams)
		},
	}

	cmd.Flags().StringP(types.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	cmd.Flags().Bool(types.FlagJSONIndet, false, "print indent result json")
	viper.BindPFlag(types.FlagNode, cmd.Flags().Lookup(types.FlagNode))

	return cmd
}
