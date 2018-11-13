package block

import (

	"errors"
	"fmt"
	"strconv"

	"github.com/QOSGroup/qbase/client"
	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/crypto"
	go_amino "github.com/tendermint/go-amino"
)

func validatorsCommand(cdc *go_amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validators [height]",
		Short: "Get validator set at given height",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) < 1 {
				return errors.New("missing height args")
			}

			h, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			height := int64(h)

			viper.Set(client.FlagTrustNode, true)
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
				PubKey crypto.PubKey
			}


			for _, validator := range validatorsRes.Validators {
				transferValidator := struct {
					Address     string
					VotingPower int64
					PubKey crypto.PubKey
				}{
					Address:     types.Address(validator.Address).String(),
					VotingPower: validator.VotingPower,
					PubKey: validator.PubKey,
				}

				transferValidators = append(transferValidators, transferValidator)
			}

			output, err := cliCtx.ToJSONIndentStr(transferValidators)
			fmt.Println(string(output))

			return nil
		},
	}

	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	viper.BindPFlag(client.FlagNode, cmd.Flags().Lookup(client.FlagNode))
	return cmd
}
