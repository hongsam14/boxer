package config

import (
	berror "boxerd/error"
	"fmt"
	"net"
	"strings"
)

const (
	MACHINE_KEYWORD  = "$machine"
	SNAPSHOT_KEYWORD = "$snapshot"
)

// VMControlConfig is a struct that holds the Commandline for the VM control
// reserved keyword:
// - $machine
// - $snapshot
type VMControlConfig struct {
	StartCmd           string `mapstructure:"start_cmd"`
	StopCmd            string `mapstructure:"stop_cmd"`
	RestoreSnapshotCmd string `mapstructure:"restore_snapshot_cmd"`
}

func (c *VMControlConfig) CheckReservedKeyword() bool {
	// check if the reserved keyword '$machine' is in the command
	if !strings.Contains(c.StartCmd, MACHINE_KEYWORD) {
		return false
	}
	// check if the reserved keyword "$machine" is in the command
	if !strings.Contains(c.StopCmd, MACHINE_KEYWORD) {
		return false
	}
	// check if the reserved keyword "$snapshot" is in the command
	if !strings.Contains(c.RestoreSnapshotCmd, SNAPSHOT_KEYWORD) ||
		!strings.Contains(c.RestoreSnapshotCmd, MACHINE_KEYWORD) {
		return false
	}
	return true
}

type VMInfoConfig struct {
	// Name is the name of the VM
	Name string `mapstructure:"name"`
	// Snapshot is the name of the snapshot
	Snapshot string `mapstructure:"snapshot"`
	IP       string `mapstructure:"ip"`
	OS       string `mapstructure:"os"`
}

type BoxerConfig struct {
	// VMInfo is the configuration for the VM
	VMInfo map[string]VMInfoConfig `mapstructure:"vm_info"`
	// VMControl is the configuration for the VM control commands
	VMControl VMControlConfig `mapstructure:"vm_control"`
}

func (bc *BoxerConfig) Validate() error {
	if !bc.VMControl.CheckReservedKeyword() {
		return berror.BoxerError{
			Code: berror.InvalidConfig,
			Msg:  "error in boxer config.Validate",
			Origin: fmt.Errorf("VM control commands must contain reserved keywords: " +
				"$machine and $snapshot"),
		}
	}
	for _, vmInfo := range bc.VMInfo {
		if vmInfo.Name == "" {
			return berror.BoxerError{
				Code:   berror.InvalidConfig,
				Msg:    "error in boxer config.Validate",
				Origin: fmt.Errorf("VM name cannot be empty"),
			}
		}
		if vmInfo.Snapshot == "" {
			return berror.BoxerError{
				Code:   berror.InvalidConfig,
				Msg:    "error in boxer config.Validate",
				Origin: fmt.Errorf("VM snapshot cannot be empty"),
			}
		}
		if vmInfo.OS == "" {
			return berror.BoxerError{
				Code:   berror.InvalidConfig,
				Msg:    "error in boxer config.Validate",
				Origin: fmt.Errorf("VM OS cannot be empty"),
			}
		}
		if vmInfo.IP == "" {
			return berror.BoxerError{
				Code:   berror.InvalidConfig,
				Msg:    "error in boxer config.Validate",
				Origin: fmt.Errorf("VM IP cannot be empty"),
			}
		}
		// check IP format
		if net.ParseIP(vmInfo.IP) == nil {
			return berror.BoxerError{
				Code:   berror.InvalidConfig,
				Msg:    "error in boxer config.Validate",
				Origin: fmt.Errorf("VM IP is not a valid IP address"),
			}
		}
	}
	return nil
}
