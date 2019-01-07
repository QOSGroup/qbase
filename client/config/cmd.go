package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
)

const (
	flagPrint = "print"
)

var configDefaults map[string]string

func init() {
	configDefaults = map[string]string{
		"chain_id": "",
		"output":   "text",
		"node":     "tcp://localhost:26657",
	}
}

func Cmd(defaultCliHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <key> [value]",
		Short: "Create or query a CLI configuration file",
		RunE:  runConfigCmd,
		Args:  cobra.RangeArgs(0, 2),
	}

	cmd.Flags().String(cli.HomeFlag, defaultCliHome, "set client's home directory for configuration")
	cmd.Flags().BoolP(flagPrint, "p", false, "print configuration value or its default if unset")
	return cmd
}

func runConfigCmd(cmd *cobra.Command, args []string) error {
	cfgFile, err := ensureConfFile(viper.GetString(cli.HomeFlag))
	if err != nil {
		return err
	}

	printFlag := viper.GetBool(flagPrint)
	if printFlag && len(args) != 1 {
		return fmt.Errorf("wrong number of arguments")
	}

	if !printFlag && len(args) == 1 {
		printFlag = true
	}

	// Load configuration
	tree, err := loadConfigFile(cfgFile)
	if err != nil {
		return err
	}

	// Print the config and exit
	if len(args) == 0 {
		s, err := tree.ToTomlString()
		if err != nil {
			return err
		}
		fmt.Print(s)
		return nil
	}

	key := args[0]
	// Get value action
	if printFlag {
		switch key {
		case "trust_node", "indent":
			fmt.Println(tree.GetDefault(key, false).(bool))
		default:
			if defaultValue, ok := configDefaults[key]; ok {
				fmt.Println(tree.GetDefault(key, defaultValue).(string))
			} else {
				s, _ := tree.Get(key).(string)
				fmt.Println(s)
			}
		}
		return nil
	}

	// Set value action
	value := args[1]
	switch key {
	case "trust_node", "indent":
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		tree.Set(key, boolVal)
	default:
		tree.Set(key, value)
	}

	// Save configuration to disk
	if err := saveConfigFile(cfgFile, tree); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "configuration saved to %s\n", cfgFile)

	return nil
}

func ensureConfFile(rootDir string) (string, error) {
	cfgPath := path.Join(rootDir, "config")
	if err := os.MkdirAll(cfgPath, os.ModePerm); err != nil {
		return "", err
	}

	return path.Join(cfgPath, "config.toml"), nil
}

func loadConfigFile(cfgFile string) (*toml.Tree, error) {
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "%s does not exist and create it\n", cfgFile)
		t, _ := toml.Load(``)
		err := saveConfigFile(cfgFile, t)
		if err != nil {
			return nil, err
		}
		return t, nil
	}

	bz, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return nil, err
	}

	tree, err := toml.LoadBytes(bz)
	if err != nil {
		return nil, err
	}

	return tree, nil
}

func saveConfigFile(cfgFile string, tree *toml.Tree) error {
	fp, err := os.OpenFile(cfgFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	_, err = tree.WriteTo(fp)
	return err
}

func errUnknownConfigKey(key string) error {
	return fmt.Errorf("unknown configuration key: %q", key)
}
