package keys

import (
	"encoding/hex"
	"encoding/json"
	"errors"
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

			prikStr := ""
			if priFile != "" {
				bz, err := ioutil.ReadFile(priFile)
				if err != nil {
					return err
				}
				prikStr = string(bz)
			} else {
				prikStr, err = utils.GetString("Enter Hex private key: ", buf)
				if err != nil {
					return err
				}
			}

			privBytes, err := readPrivateKey(prikStr)
			if err != nil {
				return err
			}

			copy(prikey[:], privBytes)

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

	cmd.Flags().String(flagPriFile, "", "import private key from acc File")

	return cmd
}

func readPrivateKey(content string) ([]byte, error) {
	privBytes, err := hex.DecodeString(content)
	if err == nil {
		return privBytes, nil
	}

	var m map[string]interface{}
	err = json.Unmarshal([]byte(content), &m)
	if err != nil {
		return nil, errors.New("content is invalid:" + err.Error())
	}

	value := ""
	ok := true
	pkInter, exists := m["privkey"]
	if exists {
		value, ok = pkInter.(string)
		if !ok {
			return nil, errors.New("content is not valid private key spec")
		}
	} else {
		return nil, errors.New("content is not valid private key spec")
	}

	return hex.DecodeString(value)
}
