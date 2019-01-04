package config

import (
	"fmt"

	"github.com/QOSGroup/qbase/types"
)

//_____________________________________________________________________

// Configuration structure for command functions that share configuration.
// For example: init, init gen-tx and testnet commands need similar input and run the same code

// Storage for init gen-tx command input parameters

const (
	defaultMinimumFees = ""
)

// BaseConfig defines the server's basic configuration
type BaseConfig struct {
	// Tx minimum fee
	MinFees string `mapstructure:"minimum_fees"`
}

// Config defines the server's top level configuration
type Config struct {
	BaseConfig `mapstructure:",squash"`
}

// SetMinimumFee sets the minimum fee.
func (c *Config) SetMinimumFees(fees types.BaseCoins) { c.MinFees = fees.String() }

// SetMinimumFee sets the minimum fee.
func (c *Config) MinimumFees() types.BaseCoins {
	fees, err := types.ParseCoins(c.MinFees)
	if err != nil {
		panic(fmt.Sprintf("invalid minimum fees: %v", err))
	}
	return fees
}

// DefaultConfig returns server's default configuration.
func DefaultConfig() *Config { return &Config{BaseConfig{MinFees: defaultMinimumFees}} }
