package keys

import (
	"encoding/hex"
	"fmt"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"io/ioutil"
	"os"
	"strings"

	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/types"
	"github.com/spf13/cobra"
	go_amino "github.com/tendermint/go-amino"
)

const (
	flagPubkey = "pubkey"
)

func exportCommand(cdc *go_amino.Codec) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "export [name]",
		Args:  cobra.ExactArgs(1),
		Short: "export key for the given name",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)

			name := args[0]
			kb, err := GetKeyBase(ctx)
			if err != nil {
				return err
			}

			info, err := kb.Get(name)
			if err != nil {
				return err
			}

			var eInfo exportInfo
			onlyExportPubkey := viper.GetBool(flagPubkey)

			if onlyExportPubkey {

				pk := ""
				if info.GetPubKey() != nil {
					pk = types.MustAccPubKeyString(info.GetPubKey())
				}

				eInfo = exportInfo{
					Name:    name,
					Address: info.GetAddress(),
					Pubkey:  pk,
				}
			} else {
				passwd, err := GetPassphrase(ctx, name)
				if err != nil {
					return err
				}
				priv, err := kb.ExportPrivateKeyObject(name, passwd)
				if err != nil {
					return err
				}

				bz := priv.(ed25519.PrivKeyEd25519)

				eInfo = exportInfo{
					Name:    name,
					Address: info.GetAddress(),
					Pubkey:  types.MustAccPubKeyString(priv.PubKey()),
					Privkey: strings.ToUpper(hex.EncodeToString(bz[:])),
				}
			}

			accFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("%s_*.acc", name))
			if err != nil {
				return err
			}
			defer accFile.Close()

			bz, err := ctx.JSONResult(eInfo)
			if err != nil {
				return err
			}

			_, err = accFile.Write(bz)

			fmt.Println("**Important** Don't leak your private key information to others.")
			fmt.Println("Please keep your private key safely, otherwise your account will be attacked.")
			fmt.Println("result file: ", accFile.Name())

			return err
		},
	}

	cmd.Flags().Bool(flagPubkey, false, "only export public key")

	return cmd
}

type exportInfo struct {
	Name    string           `json:"name"`
	Address types.AccAddress `json:"address"`
	Pubkey  string           `json:"pubkey"`
	Privkey string           `json:"privkey"`
}
