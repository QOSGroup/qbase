package keys

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tyler-smith/go-bip39"
)

const (
	mnemonicEntropySize = 256
)

func mnemonicKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mnemonic",
		Short: "Compute the bip39 mnemonic",
		Long:  "Create a bip39 mnemonic, sometimes called a seed phrase, by reading from the system entropy.",
		RunE:  runMnemonicCmd,
	}
	return cmd
}

func runMnemonicCmd(cmd *cobra.Command, args []string) error {

	var entropySeed []byte
	var err error
	entropySeed, err = bip39.NewEntropy(mnemonicEntropySize)
	if err != nil {
		return err
	}

	mnemonic, err := bip39.NewMnemonic(entropySeed[:])
	if err != nil {
		return err
	}

	fmt.Println(mnemonic)

	return nil
}
