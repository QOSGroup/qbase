package keys

import (
	"fmt"

	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/client/utils"
	"github.com/spf13/cobra"
	go_amino "github.com/tendermint/go-amino"
)

func deleteKeyCommand(cdc *go_amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete the given key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			return runDeleteCmd(cliCtx, cmd, args)
		},
	}
	return cmd
}

func runDeleteCmd(ctx context.CLIContext, cmd *cobra.Command, args []string) error {
	name := args[0]

	kb, err := GetKeyBase(ctx)
	if err != nil {
		return err
	}

	_, err = kb.Get(name)
	if err != nil {
		return err
	}

	buf := utils.BufferStdin()
	oldpass, err := utils.GetPassword(
		"DANGER - enter password to permanently delete key:", buf)
	if err != nil {
		return err
	}

	err = kb.Delete(name, oldpass)
	if err != nil {
		return err
	}
	fmt.Println("Password deleted forever (uh oh!)")
	return nil
}
