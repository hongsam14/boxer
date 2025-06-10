package vmcontroller

import (
	"boxerd/config"
	berror "boxerd/error"
	"boxerd/internal/vmcontroller/exec"
	"boxerd/vmstate"
	"fmt"
	"os"
	"strings"
)

// VMController is an interface that defines the methods for controlling a VM.
// It provides methods to start, stop, and restore snapshots of the VM.
// The VMController uses a VMContext to manage the state and information of the VM.
type VMController interface {
	// StartVM starts the VM with the given context.
	StartVM(vctx *VMContext) error
	// StopVM stops the VM with the given context.
	StopVM(vctx *VMContext) error
	// RestoreSnapshot restores the snapshot of the VM with the given context.
	RestoreSnapshot(vctx *VMContext) error
}

type vmController struct {
	vmControl *config.VMControlConfig
	vmPolicy  *config.VMControlPolicyConfig
	mux       *exec.PaddedMutex

	fdin  *os.File // file descriptor for stdin, used for executing commands
	fdout *os.File // file descriptor for stdout, used for executing commands
}

// NewVMController creates a new VMController with the given VMControlConfig.
// The VMController is responsible for executing the commands and managing the state of the VM.
// It uses the VMControlConfig to execute commands for starting, stopping, and restoring snapshots of the VM.
// It initializes the padded mutex to prevent concurrent execution of VM control commands.
// It also uses the VMContext to manage the state and information of the VM.
func NewVMController(
	fdin *os.File,
	fdout *os.File,
	vmControlConfig *config.VMControlConfig,
	vmPolicy *config.VMControlPolicyConfig) VMController {

	return &vmController{
		fdin:      fdin,
		fdout:     fdout,
		vmControl: vmControlConfig,
		vmPolicy:  vmPolicy,
		mux:       exec.InitPaddedMutex(vmPolicy.IntervalSec),
	}
}

// replaceReservedKeyword replaces the reserved keywords in the command with the actual values from the VMContext.
// It replaces the $machine keyword with the name of the VM and the $snapshot keyword with the name of the snapshot.
func (vc *vmController) replaceReservedKeyword(command string, vctx *VMContext) (argvs []string) {
	// replace the reserved keywords in the command
	command = strings.ReplaceAll(command, config.MACHINE_KEYWORD, vctx.Machine())
	command = strings.ReplaceAll(command, config.SNAPSHOT_KEYWORD, vctx.Snapshot())
	// split the command into arguments
	return strings.Split(command, " ")
}

// StartVM starts the VM with the given context.
// It checks if the VM is in a stopped state before executing the start command.
// It sets the VM state to OFFLINE after starting the VM.
func (vc *vmController) StartVM(vctx *VMContext) (err error) {
	if vctx.State() != vmstate.STOPPED {
		return berror.BoxerError{
			Code:   berror.InvalidState,
			Msg:    "error while vmcontroller.StartVM",
			Origin: fmt.Errorf("VM is not a stopped state. current state: %s, expected: %s", vctx.State(), vmstate.STOPPED),
		}
	}

	// create the arguments for the start command by replacing reserved keywords
	argv := vc.replaceReservedKeyword(vc.vmControl.StartCmd, vctx)
	if len(argv) == 0 {
		return berror.BoxerError{
			Code:   berror.InvalidArgument,
			Msg:    "error while vmcontroller.StartVM",
			Origin: fmt.Errorf("start command is empty after replacing reserved keywords %v", vc.vmControl.StartCmd),
		}
	}

	// lock the padded mutex to prevent concurrent execution of vm control commands
	vc.mux.Lock()
	defer vc.mux.Release()

	// Execute the start command
	promise, err := exec.Run(vc.fdin, vc.fdout, argv[0], argv[1:]...)
	if err != nil {
		return berror.BoxerError{
			Code:   berror.SystemError,
			Msg:    "error while vmcontroller.StartVM",
			Origin: fmt.Errorf("failed to execute start command %v: %w", argv, err),
		}
	}
	// Wait for the command to finish
	exitCode, err := promise.Wait()
	// check wait result
	if err != nil {
		// change the vm state to error state if the command failed
		vctx.setState(vmstate.ERROR)
		return berror.BoxerError{
			Code:   berror.SystemError,
			Msg:    "error while vmcontroller.StartVM",
			Origin: fmt.Errorf("error while waiting for start command to finish: %w", err),
		}
	}
	if exitCode != 0 {
		// change the vm state to error state if the command failed
		vctx.setState(vmstate.ERROR)
		return berror.BoxerError{
			Code:   berror.SystemError,
			Msg:    "error while vmcontroller.StartVM",
			Origin: fmt.Errorf("start command exited with non-zero exit code %d", exitCode),
		}
	}
	// Set the VM state to OFFLINE
	vctx.setState(vmstate.OFFLINE)
	return nil
}

// StopVM stops the VM with the given context.
// It checks if the VM is in an active state before executing the stop command.
// It sets the VM state to STOPPED after stopping the VM.
func (vc *vmController) StopVM(vctx *VMContext) (err error) {
	if vctx.State() < vmstate.OFFLINE {
		return berror.BoxerError{
			Code:   berror.InvalidState,
			Msg:    "error while vmcontroller.StopVM",
			Origin: fmt.Errorf("VM is not in an active state. current state: %s, expected: bigger than %s", vctx.State(), vmstate.OFFLINE),
		}
	}

	// create the arguments for the stop command by replacing reserved keywords
	argv := vc.replaceReservedKeyword(vc.vmControl.StopCmd, vctx)
	if len(argv) == 0 {
		return berror.BoxerError{
			Code:   berror.InvalidArgument,
			Msg:    "error while vmcontroller.StopVM",
			Origin: fmt.Errorf("stop command is empty after replacing reserved keywords %v", vc.vmControl.StopCmd),
		}
	}
	// lock the padded mutex to prevent concurrent execution of vm control commands
	vc.mux.Lock()
	defer vc.mux.Release()
	// Execute the stop command
	promise, err := exec.Run(vc.fdin, vc.fdout, argv[0], argv[1:]...)
	if err != nil {
		return berror.BoxerError{
			Code:   berror.SystemError,
			Msg:    "error while vmcontroller.StopVM",
			Origin: fmt.Errorf("failed to execute stop command %v: %w", argv, err),
		}
	}
	// Wait for the command to finish
	exitCode, err := promise.Wait()
	if err != nil {
		return berror.BoxerError{
			Code:   berror.SystemError,
			Msg:    "error while vmcontroller.StopVM",
			Origin: fmt.Errorf("error while waiting for stop command to finish: %w", err),
		}
	}
	if exitCode != 0 {
		return berror.BoxerError{
			Code:   berror.SystemError,
			Msg:    "error while vmcontroller.StopVM",
			Origin: fmt.Errorf("stop command exited with non-zero exit code %d", exitCode),
		}
	}
	vctx.setState(vmstate.STOPPED)
	return nil
}

// RestoreSnapshot restores the snapshot of the VM with the given context.
// It checks if the VM is in a stopped state before executing the restore snapshot command.
// It sets the VM state to STOPPED after restoring the snapshot.
func (vc *vmController) RestoreSnapshot(vctx *VMContext) (err error) {
	if vctx.State() != vmstate.STOPPED {
		return berror.BoxerError{
			Code:   berror.InvalidState,
			Msg:    "error while vmcontroller.RestoreSnapshot",
			Origin: fmt.Errorf("VM is not in a stopped state. current state: %s, expected: %s", vctx.State(), vmstate.STOPPED),
		}
	}

	// create the arguments for the restore snapshot command by replacing reserved keywords
	argv := vc.replaceReservedKeyword(vc.vmControl.RestoreSnapshotCmd, vctx)
	if len(argv) == 0 {
		return berror.BoxerError{
			Code:   berror.InvalidArgument,
			Msg:    "error while vmcontroller.RestoreSnapshot",
			Origin: fmt.Errorf("restore snapshot command is empty after replacing reserved keywords %v", vc.vmControl.RestoreSnapshotCmd),
		}
	}
	// lock the padded mutex to prevent concurrent execution of vm control commands
	vc.mux.Lock()
	defer vc.mux.Release()
	// Execute the restore snapshot command
	promise, err := exec.Run(vc.fdin, vc.fdout, argv[0], argv[1:]...)
	if err != nil {
		return berror.BoxerError{
			Code:   berror.SystemError,
			Msg:    "error while vmcontroller.RestoreSnapshot",
			Origin: fmt.Errorf("failed to execute restore snapshot command %v: %w", argv, err),
		}
	}
	vctx.setState(vmstate.RESTORING)
	// Wait for the command to finish
	exitCode, err := promise.Wait()
	if err != nil {
		return berror.BoxerError{
			Code:   berror.SystemError,
			Msg:    "error while vmcontroller.RestoreSnapshot",
			Origin: fmt.Errorf("error while waiting for restore snapshot command to finish: %w", err),
		}
	}
	if exitCode != 0 {
		return berror.BoxerError{
			Code:   berror.SystemError,
			Msg:    "error while vmcontroller.RestoreSnapshot",
			Origin: fmt.Errorf("restore snapshot command exited with non-zero exit code %d", exitCode),
		}
	}
	// Set the VM state to STOPPED after restoring snapshot
	vctx.setState(vmstate.STOPPED)
	return nil
}
