package keys

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"

	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/client/utils"
	"github.com/spf13/cobra"
	go_amino "github.com/tendermint/go-amino"
)

const (
	flagPriFile = "file"
)

func importCommand(cdc *go_amino.Codec) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "import [name]",
		Short: "Interactive command to import a new private key, encrypt it, and save to disk",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)

			name := args[0]
			kb, err := GetKeyBase(ctx)
			if err != nil {
				return err
			}

			_, err = kb.Get(name)
			if err == nil {
				return fmt.Errorf("name: %s already exsits", name)
			}

			buf := utils.BufferStdin()

			var prikey ed25519.PrivKeyEd25519
			priFile := viper.GetString(flagPriFile)
			if priFile != "" {
				//import from CA PRI FILE

				bz, err := ioutil.ReadFile(priFile)
				if err != nil {
					return err
				}

				err = ctx.Codec.UnmarshalJSON(bz, &prikey)
				if err != nil {
					return err
				}
			} else {

				prikStr, err := utils.GetString("Enter ed25519 private key: ", buf)
				if err != nil {
					return err
				}

				privBytes, err := base64.StdEncoding.DecodeString(prikStr)
				if err != nil {
					return err
				}

				copy(prikey[:], privBytes)
			}

			encryptPassword, err := utils.GetCheckPassword(
				"> Enter a passphrase for your key:",
				"> Repeat the passphrase:", buf)
			if err != nil {
				return err
			}

			_, err = kb.CreateImportInfo(name, encryptPassword, prikey)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().String(flagPriFile, "", "import private key from CA Pri File")

	return cmd
}
