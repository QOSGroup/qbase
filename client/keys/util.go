package keys

import (
	"fmt"
	"path/filepath"

	"github.com/QOSGroup/qbase/client"
	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/keys"
	"github.com/QOSGroup/qbase/types"
	"github.com/spf13/viper"

	"github.com/tendermint/tmlibs/cli"

	dbm "github.com/tendermint/tendermint/libs/db"
)

// KeyDBName is the directory under root where we store the keys
const KeyDBName = "keys"

// keybase is used to make GetKeyBase a singleton
var keybase keys.Keybase

type bechKeyOutFn func(keyInfo keys.Info) (KeyOutput, error)

// GetKeyInfo returns key info for a given name. An error is returned if the
// keybase cannot be retrieved or getting the info fails.
func GetKeyInfo(ctx context.CLIContext, name string) (keys.Info, error) {
	keybase, err := GetKeyBase(ctx)
	if err != nil {
		return nil, err
	}

	return keybase.Get(name)
}

// GetPassphrase returns a passphrase for a given name. It will first retrieve
// the key info for that name if the type is local, it'll fetch input from
// STDIN. Otherwise, an empty passphrase is returned. An error is returned if
// the key info cannot be fetched or reading from STDIN fails.
func GetPassphrase(ctx context.CLIContext, name string) (string, error) {
	var passphrase string

	keyInfo, err := GetKeyInfo(ctx, name)
	if err != nil {
		return passphrase, err
	}

	// we only need a passphrase for locally stored keys
	// TODO: (ref: #864) address security concerns
	if keyInfo.GetType() == keys.TypeLocal {
		passphrase, err = ReadPassphraseFromStdin(name)
		if err != nil {
			return passphrase, err
		}
	}

	return passphrase, nil
}

// ReadPassphraseFromStdin attempts to read a passphrase from STDIN return an
// error upon failure.
func ReadPassphraseFromStdin(name string) (string, error) {
	buf := client.BufferStdin()
	prompt := fmt.Sprintf("Password to sign with '%s':", name)

	passphrase, err := client.GetPassword(prompt, buf)
	if err != nil {
		return passphrase, fmt.Errorf("Error reading passphrase: %v", err)
	}

	return passphrase, nil
}

// TODO make keybase take a database not load from the directory

// GetKeyBase initializes a read-only KeyBase based on the configuration.
func GetKeyBase(ctx context.CLIContext) (keys.Keybase, error) {
	rootDir := viper.GetString(cli.HomeFlag)
	return GetKeyBaseFromDir(ctx, rootDir)
}

// GetKeyBaseFromDir initializes a read-only keybase at a particular dir.
func GetKeyBaseFromDir(ctx context.CLIContext, rootDir string) (keys.Keybase, error) {
	if keybase.IsNil() {
		db, err := dbm.NewGoLevelDB(KeyDBName, filepath.Join(rootDir, "keys"))
		if err != nil {
			return keys.Keybase{}, err
		}
		keybase = keys.New(db, ctx.Codec)
	}
	return keybase, nil
}

// used for outputting keys.Info over REST
type KeyOutput struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Address string `json:"address"`
	PubKey  string `json:"pub_key"`
	Seed    string `json:"seed,omitempty"`
}

// create a list of KeyOutput in bech32 format
func Bech32KeysOutput(infos []keys.Info) ([]KeyOutput, error) {
	kos := make([]KeyOutput, len(infos))
	for i, info := range infos {
		ko, err := Bech32KeyOutput(info)
		if err != nil {
			return nil, err
		}
		kos[i] = ko
	}
	return kos, nil
}

// create a KeyOutput in bech32 format
func Bech32KeyOutput(info keys.Info) (KeyOutput, error) {
	accAddr := types.Address(info.GetPubKey().Address().Bytes())

	return KeyOutput{
		Name:    info.GetName(),
		Type:    info.GetType().String(),
		Address: accAddr.String(),
		PubKey:  fmt.Sprintf("%s", info.GetPubKey()),
	}, nil
}

func printKeyInfo(ctx context.CLIContext, keyInfo keys.Info, bechKeyOut bechKeyOutFn) {
	ko, err := bechKeyOut(keyInfo)
	if err != nil {
		panic(err)
	}

	switch viper.Get(cli.OutputFlag) {
	case "text":
		fmt.Printf("NAME:\tTYPE:\tADDRESS:\t\t\t\t\t\tPUBKEY:\n")
		printKeyOutput(ko)
	case "json":
		out, err := ctx.Codec.MarshalJSON(ko)
		if err != nil {
			panic(err)
		}

		fmt.Println(string(out))
	default:
		fmt.Printf("NAME:\tTYPE:\tADDRESS:\t\t\t\t\t\tPUBKEY:\n")
		printKeyOutput(ko)
	}
}

func printInfos(ctx context.CLIContext, infos []keys.Info) {
	kos, err := Bech32KeysOutput(infos)
	if err != nil {
		panic(err)
	}
	switch viper.Get(cli.OutputFlag) {
	case "text":
		fmt.Printf("NAME:\tTYPE:\tADDRESS:\t\t\t\t\t\tPUBKEY:\n")
		for _, ko := range kos {
			printKeyOutput(ko)
		}
	case "json":
		out, err := ctx.Codec.MarshalJSON(kos)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(out))
	default:
		fmt.Printf("NAME:\tTYPE:\tADDRESS:\t\t\t\t\t\tPUBKEY:\n")
		for _, ko := range kos {
			printKeyOutput(ko)
		}

	}
}

func printKeyOutput(ko KeyOutput) {
	fmt.Printf("%s\t%s\t%s\t%s\n", ko.Name, ko.Type, ko.Address, ko.PubKey)
}

func printKeyAddress(info keys.Info, bechKeyOut bechKeyOutFn) {
	ko, err := bechKeyOut(info)
	if err != nil {
		panic(err)
	}

	fmt.Println(ko.Address)
}

func printPubKey(info keys.Info, bechKeyOut bechKeyOutFn) {
	ko, err := bechKeyOut(info)
	if err != nil {
		panic(err)
	}

	fmt.Println(ko.PubKey)
}
