package keys

import (
	"fmt"

	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/client/utils"
	"github.com/QOSGroup/qbase/keys/hd"
	"github.com/spf13/cobra"

	go_amino "github.com/tendermint/go-amino"
	"github.com/tyler-smith/go-bip39"
)

const (
	flagNewDefault = "default"
	flagBIP44Path  = "bip44-path"
)

func newKeyCommand(cdc *go_amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new",
		Short: "Interactive command to derive a new private key, encrypt it, and save to disk",
		Long: `Derive a new private key using an interactive command that will prompt you for each input.
Optionally specify a bip39 mnemonic, a bip39 passphrase to further secure the mnemonic,
and a bip32 HD path to derive a specific account. The key will be stored under the given name
and encrypted with the given password. The only input that is required is the encryption password.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			return runNewCmd(cliCtx, cmd, args)
		},
	}
	cmd.Flags().Bool(flagNewDefault, false, "Skip the prompts and just use the default values for everything")
	cmd.Flags().String(flagBIP44Path, "44'/118'/0'/0/0", "BIP44 path from which to derive a private key")
	return cmd
}

/*
input
	- bip44 path
	- bip39 mnemonic
	- local encryption password
output
	- armor encrypted private key (saved to file)
*/
func runNewCmd(ctx context.CLIContext, cmd *cobra.Command, args []string) error {
	name := args[0]
	kb, err := GetKeyBase(ctx)
	if err != nil {
		return err
	}

	buf := utils.BufferStdin()

	_, err = kb.Get(name)
	if err == nil {
		// account exists, ask for user confirmation
		if response, err := utils.GetConfirmation(
			fmt.Sprintf("> override the existing name %s", name), buf); err != nil || !response {
			return err
		}
	}

	flags := cmd.Flags()
	useDefaults, _ := flags.GetBool(flagNewDefault)
	bipFlag := flags.Lookup(flagBIP44Path)

	bip44Params, err := getBIP44ParamsAndPath(bipFlag.Value.String(), bipFlag.Changed || useDefaults)
	if err != nil {
		return err
	}

	var mnemonic string

	if !useDefaults {
		mnemonic, err = utils.GetString("Enter your bip39 mnemonic, or hit enter to generate one.", buf)
		if err != nil {
			return err
		}
	}

	if len(mnemonic) == 0 {
		// read entropy seed straight from crypto.Rand and convert to mnemonic
		entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
		if err != nil {
			return err
		}

		mnemonic, err = bip39.NewMnemonic(entropySeed[:])
		if err != nil {
			return err
		}
	}

	// get the encryption password
	encryptPassword, err := utils.GetCheckPassword(
		"> Enter a passphrase to encrypt your key to disk(password must be at least 8 characters):",
		"> Repeat the passphrase:", buf)
	if err != nil {
		return err
	}

	info, err := kb.Derive(name, mnemonic, encryptPassword, bip44Params.String())
	if err != nil {
		return err
	}

	_ = info
	return nil
}

func getBIP44ParamsAndPath(path string, flagSet bool) (*hd.BIP44Params, error) {
	buf := utils.BufferStdin()
	bip44Path := path

	// if it wasn't set in the flag, give it a chance to overide interactively
	if !flagSet {
		var err error

		printStep()

		bip44Path, err = utils.GetString(fmt.Sprintf("Enter your bip44 path. Default is %s\n", path), buf)
		if err != nil {
			return nil, err
		}

		if len(bip44Path) == 0 {
			bip44Path = path
		}
	}

	bip44params, err := hd.NewParamsFromPath(bip44Path)
	if err != nil {
		return nil, err
	}

	return bip44params, nil
}

func printPrefixed(msg string) {
	fmt.Printf("> %s\n", msg)
}

func printStep() {
	printPrefixed("-------------------------------------")
}
