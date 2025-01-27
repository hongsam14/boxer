package config

import (
	berror "boxerd/error"
	"strings"

	"github.com/spf13/viper"
)

// VMControlConfig is a struct that holds the Commandline for the VM control
// reserved keyword:
// - $machine
// - $snapshot
type VMControlConfig struct {
	StartCmd   string `mapstructure:"start_cmd"`
	StopCmd    string `mapstructure:"stop_cmd"`
	RestoreCmd string `mapstructure:"restore_cmd"`
}

func (c *VMControlConfig) CheckReservedKeyword() bool {
	if !strings.ContainsAny(c.StartCmd, "$machine") {
		return false
	}
	if !strings.ContainsAny(c.StopCmd, "$machine") {
		return false
	}
	if !strings.ContainsAny(c.RestoreCmd, "$snapshot") ||
		!strings.ContainsAny(c.RestoreCmd, "$machine") {
		return false
	}
	return true
}

type Config struct {
	VMControl VMControlConfig `mapstructure:"vm_control"`
}

func LoadConfig() (*Config, error) {
	var cfg *Config

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/boxerd")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, berror.BoxerError{
			Code:   berror.InvalidConfig,
			Origin: err,
			Msg:    "error while reading config",
		}
	}
	cfg = new(Config)
	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, berror.BoxerError{
			Code:   berror.InvalidConfig,
			Origin: err,
			Msg:    "error while unmarshalling config",
		}
	}
	//validate the config
	if !cfg.VMControl.CheckReservedKeyword() {
		return nil, berror.BoxerError{
			Code:   berror.InvalidConfig,
			Origin: nil,
			Msg:    "error while validating config",
		}
	}
	return cfg, nil
}
