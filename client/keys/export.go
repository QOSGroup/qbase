package keys

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/types"
	"github.com/spf13/cobra"
	go_amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"
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
				eInfo = exportInfo{
					Name:    name,
					Address: info.GetAddress(),
					Pubkey:  info.GetPubKey(),
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

				eInfo = exportInfo{
					Name:    name,
					Address: info.GetAddress(),
					Pubkey:  priv.PubKey(),
					Privkey: priv,
				}
			}

			fmt.Println("**Important** Don't leak your private key information to others.")
			fmt.Println("Please keep your private key safely, otherwise your account will be attacked.")
			fmt.Println()

			return ctx.PrintResult(eInfo)
		},
	}

	cmd.Flags().Bool(flagPubkey, false, "only export public key")

	return cmd
}

type exportInfo struct {
	Name    string         `json:'name'`
	Address types.Address  `json:"address"`
	Pubkey  crypto.PubKey  `json:"pubkey"`
	Privkey crypto.PrivKey `json:"privkey"`
}
