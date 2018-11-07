package keys

import (
	"fmt"

	"github.com/QOSGroup/qbase/client"
	"github.com/QOSGroup/qbase/client/context"
	"github.com/spf13/cobra"
	go_amino "github.com/tendermint/go-amino"
)

func updateKeyCommand(cdc *go_amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Change the password used to protect private key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			return runUpdateCmd(cliCtx, cmd, args)
		},
	}
	return cmd
}

func runUpdateCmd(ctx context.CLIContext, cmd *cobra.Command, args []string) error {
	name := args[0]

	buf := client.BufferStdin()
	kb, err := GetKeyBase(ctx)
	if err != nil {
		return err
	}
	oldpass, err := client.GetPassword(
		"Enter the current passphrase:", buf)
	if err != nil {
		return err
	}

	getNewpass := func() (string, error) {
		return client.GetCheckPassword(
			"Enter the new passphrase:",
			"Repeat the new passphrase:", buf)
	}

	err = kb.Update(name, oldpass, getNewpass)
	if err != nil {
		return err
	}
	fmt.Println("Password successfully updated!")
	return nil
}
