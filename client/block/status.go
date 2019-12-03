package block

import (
	"errors"
	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/client/types"
	btypes "github.com/QOSGroup/qbase/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/p2p"
	core_types "github.com/tendermint/tendermint/rpc/core/types"

	go_amino "github.com/tendermint/go-amino"
)

func statusCommand(cdc *go_amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Query remote node for status",
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.Set(types.FlagTrustNode, true)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			node, err := cliCtx.GetNode()
			if err != nil {
				return err
			}

			status, err := node.Status()
			if err != nil {
				return err
			}

			if status == nil {
				return errors.New("query status return empty response")
			}

			var sdi statusDisplayInfo
			sdi.NodeInfo = status.NodeInfo
			sdi.SyncInfo = status.SyncInfo

			consPubKey, _ := btypes.ConsensusPubKeyString(status.ValidatorInfo.PubKey)

			sdi.ValidatorInfo = consValidatorInfo{
				Address:     btypes.ConsAddress(status.ValidatorInfo.Address.Bytes()).String(),
				PubKey:      consPubKey,
				VotingPower: status.ValidatorInfo.VotingPower,
			}

			return cliCtx.PrintResult(sdi)
		},
	}

	cmd.Flags().StringP(types.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	cmd.Flags().Bool(types.FlagJSONIndet, false, "print indent result json")
	viper.BindPFlag(types.FlagNode, cmd.Flags().Lookup(types.FlagNode))

	return cmd
}

type statusDisplayInfo struct {
	NodeInfo      p2p.DefaultNodeInfo `json:"node_info"`
	SyncInfo      core_types.SyncInfo `json:"sync_info"`
	ValidatorInfo consValidatorInfo   `json:"validator_info"`
}

type consValidatorInfo struct {
	Address     string `json:"address"`
	PubKey      string `json:"pub_key"`
	VotingPower int64  `json:"voting_power"`
}
