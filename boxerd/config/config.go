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

// VMPolicyConfig is a struct that holds the policy configuration for the VMControl
type VMControlPolicyConfig struct {
	IntervalSec     uint `mapstructure:"interval"`          // Interval is the interval in seconds for the VM control commands
	TimeoutSec      uint `mapstructure:"timeout"`           // Timeout is the timeout in seconds for the VM control commands
	MaxVMOperations uint `mapstructure:"max_vm_operations"` // MaxVMOperations is the maximum number of VM operations that can be performed in parallel
}

func (c *VMControlPolicyConfig) Validate() error {
	if c.IntervalSec == 0 {
		return berror.BoxerError{
			Code:   berror.InvalidConfig,
			Msg:    "error in VMControlPolicyConfig Validate",
			Origin: fmt.Errorf("VM control policy interval cannot be zero"),
		}
	}
	if c.TimeoutSec == 0 {
		return berror.BoxerError{
			Code:   berror.InvalidConfig,
			Msg:    "error in VMControlPolicyConfig Validate",
			Origin: fmt.Errorf("VM control policy timeout cannot be zero"),
		}
	}
	if c.MaxVMOperations == 0 {
		return berror.BoxerError{
			Code:   berror.InvalidConfig,
			Msg:    "error in VMControlPolicyConfig Validate",
			Origin: fmt.Errorf("VM control policy max VM operations cannot be zero"),
		}
	}
	return nil
}

// VMInfoConfig is a struct that holds the information of the VM
// It includes the name of the VM, the snapshot name, the IP address, the OS type,
// and the group of the VM.
type VMInfoConfig struct {
	// Name is the name of the VM
	Name string `mapstructure:"name"`
	// Snapshot is the name of the snapshot
	Snapshot string `mapstructure:"snapshot"`
	IP       string `mapstructure:"ip"`
	OS       string `mapstructure:"os"`
	Group    string `mapstructure:"group"` // Group is the group of the VM, used for grouping VMs in the UI
}

func (v *VMInfoConfig) Validate() error {
	if v.Name == "" {
		return berror.BoxerError{
			Code:   berror.InvalidConfig,
			Msg:    "error in VMInfoConfig Validate",
			Origin: fmt.Errorf("VM name cannot be empty"),
		}
	}
	if v.Snapshot == "" {
		return berror.BoxerError{
			Code:   berror.InvalidConfig,
			Msg:    "error in VMInfoConfig Validate",
			Origin: fmt.Errorf("VM snapshot cannot be empty"),
		}
	}
	if v.OS == "" {
		return berror.BoxerError{
			Code:   berror.InvalidConfig,
			Msg:    "error in VMInfoConfig Validate",
			Origin: fmt.Errorf("VM OS cannot be empty"),
		}
	}
	if v.Group == "" {
		return berror.BoxerError{
			Code:   berror.InvalidConfig,
			Msg:    "error in VMInfoConfig Validate",
			Origin: fmt.Errorf("VM group cannot be empty"),
		}
	}
	if v.IP == "" {
		return berror.BoxerError{
			Code:   berror.InvalidConfig,
			Msg:    "error in VMInfoConfig Validate",
			Origin: fmt.Errorf("VM IP cannot be empty"),
		}
	}
	if net.ParseIP(v.IP) == nil {
		return berror.BoxerError{
			Code:   berror.InvalidConfig,
			Msg:    "error in VMInfoConfig Validate",
			Origin: fmt.Errorf("VM IP is not a valid IP address"),
		}
	}
	return nil
}

type BoxerConfig struct {
	// VMInfo is the configuration for the VM
	VMInfo map[string]VMInfoConfig `mapstructure:"vm_info"`
	// VMControl is the configuration for the VM control commands
	VMControl VMControlConfig `mapstructure:"vm_control"`
	// VMControlPolicy is the configuration for the VM control policy
	VMControlPolicy VMControlPolicyConfig `mapstructure:"vm_control_policy"`
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
		if err := vmInfo.Validate(); err != nil {
			return berror.BoxerError{
				Code:   berror.InvalidConfig,
				Msg:    "error in boxer config.Validate",
				Origin: fmt.Errorf("invalid VM info config for VM %s: %w", vmInfo.Name, err),
			}
		}
	}
	return bc.VMControlPolicy.Validate()
}
