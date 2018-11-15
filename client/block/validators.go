package block

import (
	"fmt"
	"strconv"

	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/client/types"
	btypes "github.com/QOSGroup/qbase/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	go_amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"
)

func validatorsCommand(cdc *go_amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validators [height]",
		Short: "Get validator set at given height",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			h, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			height := int64(h)

			viper.Set(types.FlagTrustNode, true)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			node, err := cliCtx.GetNode()
			if err != nil {
				return err
			}

			validatorsRes, err := node.Validators(&height)
			if err != nil {
				return err
			}

			var transferValidators []struct {
				Address     string
				VotingPower int64
				PubKey      crypto.PubKey
			}

			for _, validator := range validatorsRes.Validators {
				transferValidator := struct {
					Address     string
					VotingPower int64
					PubKey      crypto.PubKey
				}{
					Address:     btypes.Address(validator.Address).String(),
					VotingPower: validator.VotingPower,
					PubKey:      validator.PubKey,
				}

				transferValidators = append(transferValidators, transferValidator)
			}

			output, err := cliCtx.ToJSONIndentStr(transferValidators)
			fmt.Println(string(output))

			return nil
		},
	}

	cmd.Flags().StringP(types.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	viper.BindPFlag(types.FlagNode, cmd.Flags().Lookup(types.FlagNode))
	return cmd
}
