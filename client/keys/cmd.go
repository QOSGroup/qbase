package keys

import (
	"github.com/QOSGroup/qbase/client"

	"github.com/spf13/cobra"
	go_amino "github.com/tendermint/go-amino"
)

// Commands registers a sub-tree of commands to interact with
// local private key storage.
func Commands(cdc *go_amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Add or view local private keys",
		Long: `Keys allows you to manage your local keystore for tendermint.

    These keys may be in any format supported by go-crypto and can be
    used by light-clients, full nodes, or any other application that
    needs to sign with a private key.`,
	}
	cmd.AddCommand(
		mnemonicKeyCommand(),
		newKeyCommand(cdc),
		addKeyCommand(cdc),
		listKeysCmd(cdc),
		client.LineBreak,
		deleteKeyCommand(cdc),
		updateKeyCommand(cdc),
	)
	return cmd
}
