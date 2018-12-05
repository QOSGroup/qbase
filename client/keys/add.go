package keys

import (
	"fmt"

	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/client/utils"
	"github.com/QOSGroup/qbase/keys"
	"github.com/QOSGroup/qbase/keys/hd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	go_amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/cli"
)

const (
	flagRecover = "recover"
)

func addKeyCommand(cdc *go_amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name>",
		Args:  cobra.ExactArgs(1),
		Short: "Create a new key, or import from seed",
		Long: `Add a public/private key pair to the key store.
If you select --recover you can recover a key from the seed
phrase, otherwise, a new key will be generated.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			return runAddCmd(cliCtx, cmd, args)
		},
	}
	cmd.Flags().Bool(flagRecover, false, "Provide seed phrase to recover existing key instead of creating")
	return cmd
}

func runAddCmd(ctx context.CLIContext, cmd *cobra.Command, args []string) error {
	var kb keys.Keybase
	var err error
	var name, pass string

	buf := utils.BufferStdin()
	name = args[0]
	kb, err = GetKeyBase(ctx)
	if err != nil {
		return err
	}

	_, err = kb.Get(name)
	if err == nil {
		if response, err := utils.GetConfirmation(
			fmt.Sprintf("override the existing name %s", name), buf); err != nil || !response {
			return err
		}
	}

	pass, err = utils.GetCheckPassword(
		"Enter a passphrase for your key:",
		"Repeat the passphrase:", buf)
	if err != nil {
		return err
	}

	if viper.GetBool(flagRecover) {
		seed, err := utils.GetSeed(
			"Enter your recovery seed phrase:", buf)
		if err != nil {
			return err
		}
		info, err := kb.Derive(name, seed, pass, hd.FullFundraiserPath)
		if err != nil {
			return err
		}
		printCreate(ctx, info, "")
	} else {
		info, seed, err := kb.CreateEnMnemonic(name, pass)
		if err != nil {
			return err
		}
		printCreate(ctx, info, seed)
	}
	return nil
}

func printCreate(ctx context.CLIContext, info keys.Info, seed string) {
	output := viper.Get(cli.OutputFlag)
	switch output {
	case "json":
		out, err := Bech32KeyOutput(ctx, info)
		if err != nil {
			panic(err)
		}
		out.Seed = seed
		var jsonString []byte
		jsonString, err = ctx.Codec.MarshalJSONIndent(out, "", "  ")

		if err != nil {
			panic(err) // really shouldn't happen...
		}
		fmt.Println(string(jsonString))
	default:
		printKeyInfo(ctx, info, Bech32KeyOutput)

		fmt.Println("**Important** write this seed phrase in a safe place.")
		fmt.Println("It is the only way to recover your account if you ever forget your password.")
		fmt.Println()
		fmt.Println(seed)
	}
}
