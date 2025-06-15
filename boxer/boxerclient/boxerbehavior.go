package boxer

import (
	"github.com/hongsam14/boxer/internal/vmcontroller"
	"github.com/hongsam14/boxer/vmstate"
)

// BoxerOp is used to define the operation to be performed by the boxer.
type BoxerOp int

const (
	// STOP represents stopping a VM.
	STOP BoxerOp = iota
	// START represents starting a VM.
	START
	// RESTART represents restarting a VM.
	RESTORE
)

// String() returns the string representation of the BoxerOp.
func (op BoxerOp) String() string {
	switch op {
	case STOP:
		return "STOP"
	case START:
		return "START"
	case RESTORE:
		return "RESTORE"
	default:
		return "UNKNOWN"
	}
}

// ReturnCode represents the result of a boxer operation.
type ReturnCode int

const (
	NOT_INITIALIZED ReturnCode = iota
	SUCCESS
	INTERNAL_ERROR
	NOT_FOUND
	INVALID_REQUEST
	ALREADY_EXISTS
)

// String() returns the string representation of the ReturnCode.
func (rc ReturnCode) String() string {
	switch rc {
	case NOT_INITIALIZED:
		return "NOT_INITIALIZED"
	case SUCCESS:
		return "SUCCESS"
	case INTERNAL_ERROR:
		return "INTERNAL_ERROR"
	case NOT_FOUND:
		return "NOT_FOUND"
	case INVALID_REQUEST:
		return "INVALID_REQUEST"
	case ALREADY_EXISTS:
		return "ALREADY_EXISTS"
	default:
		return "UNKNOWN"
	}
}

// Box represents a virtual machine with its associated properties.
type Box interface {
	// Machine returns the name of the VM.
	Machine() string
	// Group returns the group name of the VM.
	Group() string
	// IP returns the IP address of the VM.
	IP() string
	// OS returns the operating system of the VM.
	OS() string
	// State returns the current state of the VM.
	State() vmstate.VMState
}

type box struct {
	group   string
	machine string
	ip      string
	os      string
	state   vmstate.VMState
}

// Machine returns the name of the VM.
func (b *box) Machine() string {
	return b.machine
}

func (b *box) Group() string {
	return b.group
}

// IP returns the IP address of the VM.
func (b *box) IP() string {
	return b.ip
}

// OS returns the operating system of the VM.
func (b *box) OS() string {
	return b.os
}

// State returns the current state of the VM.
func (b *box) State() vmstate.VMState {
	return b.state
}

// NewBox creates a new Box instance with the provided parameters.
func NewBox(vmCtx *vmcontroller.VMContext) Box {
	return &box{
		machine: vmCtx.Machine(),
		group:   vmCtx.Group(),
		ip:      vmCtx.IP(),
		os:      vmCtx.OS(),
		state:   vmCtx.State(),
	}
}

// BoxerRequest is used to request an operation on a BoxerClient
type BoxerRequest struct {
	OP      BoxerOp
	BoxInfo Box
}

// BoxerResponse is used to return the result of an operation on a BoxerClient
type BoxerResponse struct {
	Code    ReturnCode
	BoxInfo Box
}
