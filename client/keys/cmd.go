package keys

import (
	"github.com/QOSGroup/qbase/client/types"
	"github.com/spf13/cobra"
	go_amino "github.com/tendermint/go-amino"
)

func KeysCommand(cdc *go_amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "keys management tools. Add or view local private keys",
		Long: `Keys allows you to manage your local keystore for tendermint.

    These keys may be in any format supported by go-crypto and can be
    used by light-clients, full nodes, or any other application that
    needs to sign with a private key.`,
	}
	cmd.AddCommand(
		addKeyCommand(cdc),
		listKeysCmd(cdc),
		types.LineBreak,
		deleteKeyCommand(cdc),
		updateKeyCommand(cdc),
		types.LineBreak,
		exportCommand(cdc),
		importCommand(cdc),
	)

	return cmd
}
