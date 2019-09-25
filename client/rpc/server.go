package rpc

import (
	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/client/types"
	"github.com/QOSGroup/qbase/server"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/log"
	rpcserver "github.com/tendermint/tendermint/rpc/lib/server"
	"net"
	"os"
	"time"
)

type RestServer struct {
	Mux    *mux.Router
	CliCtx context.CLIContext

	log      log.Logger
	listener net.Listener
}

type Config struct {
	MaxOpen      uint64
	ReadTimeOut  uint64
	WriteTimeOut uint64
}

func NewRestServer(cdc *amino.Codec) *RestServer {
	r := mux.NewRouter()
	ctx := context.NewCLIContext().WithCodec(cdc)
	log := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "rpc-server")
	return &RestServer{
		Mux:    r,
		CliCtx: ctx,
		log:    log,
	}
}

func (rs *RestServer) Start(listenAddr string, config Config) (err error) {
	cfg := rpcserver.DefaultConfig()
	cfg.MaxOpenConnections = int(config.MaxOpen)
	cfg.ReadTimeout = time.Duration(config.ReadTimeOut) * time.Second
	cfg.WriteTimeout = time.Duration(config.WriteTimeOut) * time.Second

	rs.listener, err = rpcserver.Listen(listenAddr, cfg)
	if err != nil {
		return
	}

	server.TrapSignal(func() {
		err := rs.listener.Close()
		rs.log.Error("error closing listener", "err", err)
	})

	rs.log.Info("Starting application REST service...")
	return rpcserver.StartHTTPServer(rs.listener, rs.Mux, rs.log, cfg)
}

func ServerCommand(cdc *amino.Codec, registerRoutesFn func(*RestServer)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rpc-server",
		Short: "Start a Local RPC Server",
		RunE: func(cmd *cobra.Command, args []string) error {

			rs := NewRestServer(cdc)
			rs.CliCtx = rs.CliCtx.WithChainID(viper.GetString(types.FlagChainID))

			registerTxsRoutes(rs.CliCtx, rs.Mux)
			registerQueryRoutes(rs.CliCtx, rs.Mux)

			registerRoutesFn(rs)

			return rs.Start(viper.GetString(types.FlagListenAddr), Config{
				MaxOpen:      uint64(viper.GetInt64(types.FlagMaxOpenConnections)),
				ReadTimeOut:  uint64(viper.GetInt64(types.FlagRPCReadTimeout)),
				WriteTimeOut: uint64(viper.GetInt64(types.FlagRPCWriteTimeout)),
			})
		},
	}

	cmd.Flags().String(types.FlagChainID, "", "Chain id")
	cmd.Flags().Bool(types.FlagJSONIndet, false, "Add indent to JSON response")
	cmd.Flags().Bool(types.FlagTrustNode, false, "Trust connected full node (don't verify proofs for responses)")
	cmd.Flags().String(types.FlagNode, "tcp://localhost:26657", "<host>:<port> to Tendermint RPC interface for this chain")

	viper.BindPFlag(types.FlagTrustNode, cmd.Flags().Lookup(types.FlagTrustNode))
	viper.BindPFlag(types.FlagNode, cmd.Flags().Lookup(types.FlagNode))

	cmd.Flags().String(types.FlagListenAddr, "tcp://localhost:9876", "The address for the server to listen on")
	cmd.Flags().Uint(types.FlagMaxOpenConnections, 1000, "The number of maximum open connections")
	cmd.Flags().Uint(types.FlagRPCReadTimeout, 10, "The RPC read timeout (in seconds)")
	cmd.Flags().Uint(types.FlagRPCWriteTimeout, 10, "The RPC write timeout (in seconds)")

	viper.BindPFlag(types.FlagListenAddr, cmd.Flags().Lookup(types.FlagListenAddr))
	viper.BindPFlag(types.FlagMaxOpenConnections, cmd.Flags().Lookup(types.FlagMaxOpenConnections))
	viper.BindPFlag(types.FlagRPCReadTimeout, cmd.Flags().Lookup(types.FlagRPCReadTimeout))
	viper.BindPFlag(types.FlagRPCWriteTimeout, cmd.Flags().Lookup(types.FlagRPCWriteTimeout))

	cmd.MarkFlagRequired(types.FlagChainID)

	return cmd
}
