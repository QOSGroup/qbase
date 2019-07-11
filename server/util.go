package server

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/QOSGroup/qbase/version"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	go_amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/common"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/cli"
	tmflags "github.com/tendermint/tendermint/libs/cli/flags"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/types"
)

// server context
type Context struct {
	Config *cfg.Config
	Logger log.Logger
}

func NewDefaultContext() *Context {
	return NewContext(
		cfg.DefaultConfig(),
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
	)
}

func NewContext(config *cfg.Config, logger log.Logger) *Context {
	return &Context{config, logger}
}

//___________________________________________________________________________________

// PersistentPreRunEFn returns a PersistentPreRunE function for cobra
// that initailizes the passed in context with a properly configured
// logger and config objecy
func PersistentPreRunEFn(context *Context) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == version.VersionCmd.Name() {
			return nil
		}
		config, err := interceptLoadConfig()
		if err != nil {
			return err
		}
		logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
		logger, err = tmflags.ParseLogLevel(config.LogLevel, logger, cfg.DefaultLogLevel())
		if err != nil {
			return err
		}
		if viper.GetBool(cli.TraceFlag) {
			logger = log.NewTracingLogger(logger)
		}
		logger = logger.With("module", "main")
		context.Config = config
		context.Logger = logger
		return nil
	}
}

// If a new config is created, change some of the default tendermint settings
func interceptLoadConfig() (conf *cfg.Config, err error) {
	tmpConf := cfg.DefaultConfig()
	err = viper.Unmarshal(tmpConf)
	if err != nil {
		panic(err)
	}
	rootDir := tmpConf.RootDir
	configFilePath := filepath.Join(rootDir, "config/config.toml")
	// Intercept only if the file doesn't already exist

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		// the following parse config is needed to create directories
		conf, _ = tcmd.ParseConfig()
		conf.ProfListenAddress = "localhost:6060"
		conf.P2P.RecvRate = 5120000
		conf.P2P.SendRate = 5120000
		conf.TxIndex.IndexAllTags = true
		conf.Consensus.TimeoutCommit = 5 * time.Second
		cfg.WriteConfigFile(configFilePath, conf)
		// Fall through, just so that its parsed into memory.
	}

	if conf == nil {
		conf, err = tcmd.ParseConfig()
	}

	return
}

// add server commands
func AddCommands(
	ctx *Context, cdc *go_amino.Codec,
	rootCmd *cobra.Command, appCreator AppCreator) {

	rootCmd.PersistentFlags().String("log_level", ctx.Config.LogLevel, "Log level")

	tendermintCmd := &cobra.Command{
		Use:   "tendermint",
		Short: "Tendermint subcommands",
	}

	tendermintCmd.AddCommand(
		ShowNodeIDCmd(ctx),
		ShowValidatorCmd(ctx),
		ShowAddressCmd(ctx),
	)

	rootCmd.AddCommand(
		StartCmd(ctx, appCreator),
		UnsafeResetAllCmd(ctx),
		tendermintCmd,
	)
}

//___________________________________________________________________________________

// InsertKeyJSON inserts a new JSON field/key with a given value to an existing
// JSON message. An error is returned if any serialization operation fails.
//
// NOTE: The ordering of the keys returned as the resulting JSON message is
// non-deterministic, so the client should not rely on key ordering.
func InsertKeyJSON(cdc *go_amino.Codec, baseJSON []byte, key string, value json.RawMessage) ([]byte, error) {
	var jsonMap map[string]json.RawMessage

	if err := cdc.UnmarshalJSON(baseJSON, &jsonMap); err != nil {
		return nil, err
	}

	jsonMap[key] = value
	bz, err := cdc.MarshalJSONIndent(jsonMap, "", " ")

	return json.RawMessage(bz), err
}

// https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
// TODO there must be a better way to get external IP
func ExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if skipInterface(iface) {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			ip := addrToIP(addr)
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

// TrapSignal traps SIGINT and SIGTERM and terminates the server correctly.
func TrapSignal(cleanupFunc func()) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		switch sig {
		case syscall.SIGTERM:
			defer cleanupFunc()
			os.Exit(128 + int(syscall.SIGTERM))
		case syscall.SIGINT:
			defer cleanupFunc()
			os.Exit(128 + int(syscall.SIGINT))
		}
	}()
}

func skipInterface(iface net.Interface) bool {
	if iface.Flags&net.FlagUp == 0 {
		return true // interface down
	}
	if iface.Flags&net.FlagLoopback != 0 {
		return true // loopback interface
	}
	return false
}

func addrToIP(addr net.Addr) net.IP {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	return ip
}

func SaveGenDoc(genFile string, genDoc types.GenesisDoc) error {
	if err := genDoc.ValidateAndComplete(); err != nil {
		return err
	}

	return genDoc.SaveAs(genFile)
}

// read of create the private key file for this config
func ReadOrCreatePrivValidator(privValFile, stateFile string) crypto.PubKey {
	var privValidator *privval.FilePV

	if common.FileExists(privValFile) {
		privValidator = privval.LoadFilePV(privValFile, stateFile)
	} else {
		privValidator = privval.GenFilePV(privValFile, stateFile)
		privValidator.Save()
	}

	return privValidator.GetPubKey()
}

// InitializeNodeValidatorFiles creates private validator and p2p configuration files.
func InitializeNodeValidatorFiles(
	config *cfg.Config) (nodeID string, valPubKey crypto.PubKey, err error,
) {

	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return nodeID, valPubKey, err
	}

	nodeID = string(nodeKey.ID())
	valPubKey = ReadOrCreatePrivValidator(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile())

	return nodeID, valPubKey, nil
}

func loadGenesisDoc(cdc *go_amino.Codec, genFile string) (genDoc types.GenesisDoc, err error) {
	genContents, err := ioutil.ReadFile(genFile)
	if err != nil {
		return genDoc, err
	}

	if err := cdc.UnmarshalJSON(genContents, &genDoc); err != nil {
		return genDoc, err
	}

	return genDoc, err
}
