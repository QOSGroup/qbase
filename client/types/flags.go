package types

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	FlagChainID   = "chain-id"
	FlagNode      = "node"
	FlagHeight    = "height"
	FlagNonce     = "nonce"
	FlagAsync     = "async"
	FlagTrustNode = "trust-node"
	FlagMaxGas    = "max-gas"
)

// LineBreak can be included in a command list to provide a blank line
// to help with readability
var (
	LineBreak = &cobra.Command{Run: func(*cobra.Command, []string) {}}
)

// GetCommands adds common flags to query commands
func GetCommands(cmds ...*cobra.Command) []*cobra.Command {
	for _, c := range cmds {
		c.Flags().String(FlagChainID, "", "Chain ID of tendermint node")
		c.Flags().Bool(FlagTrustNode, false, "Trust connected full node (don't verify proofs for responses)")
		c.Flags().String(FlagNode, "tcp://localhost:26657", "<host>:<port> to tendermint rpc interface for this chain")
		c.Flags().Int64(FlagHeight, 0, "block height to query, omit to get most recent provable block")
		viper.BindPFlag(FlagChainID, c.Flags().Lookup(FlagChainID))
		viper.BindPFlag(FlagNode, c.Flags().Lookup(FlagNode))
	}
	return cmds
}

// PostCommands adds common flags for commands to post tx
func PostCommands(cmds ...*cobra.Command) []*cobra.Command {
	for _, c := range cmds {
		c.Flags().Int64(FlagNonce, 0, "account nonce to sign the tx")
		c.Flags().Int64(FlagMaxGas, 0, "gas limit to set per tx")
		c.Flags().String(FlagChainID, "", "Chain ID of tendermint node")
		c.Flags().String(FlagNode, "tcp://localhost:26657", "<host>:<port> to tendermint rpc interface for this chain")
		c.Flags().Bool(FlagAsync, false, "broadcast transactions asynchronously")
		c.Flags().Bool(FlagTrustNode, true, "Trust connected full node (don't verify proofs for responses)")
		viper.BindPFlag(FlagChainID, c.Flags().Lookup(FlagChainID))
		viper.BindPFlag(FlagMaxGas, c.Flags().Lookup(FlagMaxGas))
		viper.BindPFlag(FlagNode, c.Flags().Lookup(FlagNode))
	}
	return cmds
}
