package keys

import (
	"errors"
	"fmt"
	"github.com/QOSGroup/qbase/client/context"
	types2 "github.com/QOSGroup/qbase/client/types"
	"github.com/QOSGroup/qbase/client/utils"
	"github.com/QOSGroup/qbase/keys"
	"github.com/QOSGroup/qbase/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"
	"strings"
)

var formatFlag = "ed25519-pubkey"
var privkeyFlag = "ed25519-privkey"

type keyInfoBech32 struct {
	Name    string         `json:"name"`
	Type    string         `json:"type"`
	Address string         `json:"address"`
	Pubkey  string         `json:"pubkey"`
	Privkey crypto.PrivKey `json:"privkey,omitempty"`
}

type keyInfoEd25519 struct {
	Name    string         `json:"name"`
	Type    string         `json:"type"`
	Address string         `json:"address"`
	Pubkey  crypto.PubKey  `json:"pubkey"`
	Privkey crypto.PrivKey `json:"privkey,omitempty"`
}

func showKeysCommand(cdc *amino.Codec) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "show <name or address>",
		Short: "show key detail from name or address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			v := args[0]

			var err error
			var info keys.Info

			kb, err := GetKeyBase(cliCtx)
			if err != nil {
				return err
			}

			if strings.HasPrefix(v, types.GetAddressConfig().GetBech32AccountAddrPrefix()) {
				addr, err := types.AccAddressFromBech32(v)
				if err != nil {
					return err
				}
				info, err = kb.GetByAddress(addr)
			} else {
				info, err = kb.Get(v)
			}

			if err != nil {
				return err
			}

			isUse25519Format := viper.GetBool(formatFlag)
			isShowPrivate := viper.GetBool(privkeyFlag)

			var privKey crypto.PrivKey
			if isShowPrivate {

				pass, err := utils.GetPassword("input password:", utils.BufferStdin())
				if err != nil {
					return err
				}

				privKey, err = kb.ExportPrivateKeyObject(info.GetName(), pass)
				if err != nil {
					return err
				}

				fmt.Println("**Important** Don't leak your private key information to others.")
				fmt.Println("Please keep your private key safely, otherwise your account will be attacked.")
				fmt.Println()
			}

			if isUse25519Format {
				err = cliCtx.PrintResult(keyInfoEd25519{
					Name:    info.GetName(),
					Type:    info.GetType().String(),
					Address: info.GetAddress().String(),
					Pubkey:  info.GetPubKey(),
					Privkey: privKey,
				})
			} else {
				err = cliCtx.PrintResult(keyInfoBech32{
					Name:    info.GetName(),
					Type:    info.GetType().String(),
					Address: info.GetAddress().String(),
					Pubkey:  types.MustAccPubKeyString(info.GetPubKey()),
					Privkey: privKey,
				})
			}

			return nil
		},
	}

	cmd.Flags().Bool(formatFlag, false, "show pubkey with tendermint ed25519 style.")
	cmd.Flags().Bool(privkeyFlag, false, "show private key. keep secure for use this.")
	return types2.GetCommands(cmd)[0]
}

func covertPubkeyCommand(cdc *amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "covert <bech32-pubkey>",
		Short: "covert bech32 PubKey to ed25519 style",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			addressConfig := types.GetAddressConfig()

			bech32Pubkey := args[0]

			var err error
			var pk crypto.PubKey

			if strings.HasPrefix(bech32Pubkey, addressConfig.GetBech32AccountPubPrefix()) {
				pk, err = types.GetAccPubKeyBech32(bech32Pubkey)
			} else if strings.HasPrefix(bech32Pubkey, addressConfig.GetBech32ValidatorPubPrefix()) {
				pk, err = types.GetValidatorPubKeyBech32(bech32Pubkey)
			} else if strings.HasPrefix(bech32Pubkey, addressConfig.GetBech32ConsensusPubPrefix()) {
				pk, err = types.GetConsensusPubKeyBech32(bech32Pubkey)
			} else {
				err = errors.New("invalid input bech32 pubkey")
			}

			if err != nil {
				return err
			}

			cliCtx.PrintResult(pk)
			return nil
		},
	}

	return types2.GetCommands(cmd)[0]
}
