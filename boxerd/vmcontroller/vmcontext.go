package vmcontroller

import (
	"boxerd/config"
	"boxerd/vmcontroller/vmstate"
)

type VMContext struct {
	info  config.VMInfoConfig
	state vmstate.VMState
}

// Machine returns the name of the VM.
func (vc *VMContext) Machine() string {
	return vc.info.Name
}

// Snapshot returns the name of the snapshot.
func (vc *VMContext) Snapshot() string {
	return vc.info.Snapshot
}

// IP returns the IP address of the VM.
func (vc *VMContext) IP() string {
	return vc.info.IP
}

// OS returns the operating system of the VM.
func (vc *VMContext) OS() string {
	return vc.info.OS
}

// Group returns the group name of the VM.
func (vc *VMContext) Group() string {
	return vc.info.Group
}

// State returns the current state of the VM.
func (vc *VMContext) State() vmstate.VMState {
	return vc.state
}

// SetState sets the current state of the VM.
func (vc *VMContext) setState(state vmstate.VMState) {
	vc.state = state
}

// NewVMContext creates a new VMContext with the provided VMInfoConfig.
func NewVMContext(info config.VMInfoConfig) *VMContext {
	return &VMContext{
		info:  info,
		state: vmstate.STOPPED, // Default state is STOPPED
	}
}
