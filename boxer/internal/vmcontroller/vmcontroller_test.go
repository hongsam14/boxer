package vmcontroller_test

import (
	"os"
	"testing"
	"time"

	"github.com/hongsam14/boxer/config"
	"github.com/hongsam14/boxer/internal/vmcontroller"
	"github.com/hongsam14/boxer/vmstate"
)

func TestVMController(t *testing.T) {
	vmInfo := config.VMInfoConfig{
		// TODO: Change this to a real VM name and snapshot
		Name: "sb_win10_develop_v2",
		// TODO: Change this to a real snapshot name
		Snapshot: "snapshot0",
		// TODO: Change this to a real IP address
		IP:    "127.0.0.1",
		OS:    "windows",
		Group: "test",
	}
	vmControlConfig := config.VMControlConfig{
		StartCmd:           "VBoxManage startvm $machine",
		StopCmd:            "VBoxManage controlvm $machine poweroff",
		RestoreSnapshotCmd: "VBoxManage snapshot $machine restore $snapshot",
	}
	if !vmControlConfig.CheckReservedKeyword() {
		t.Errorf("CheckReservedKeyword failed")
		return
	}
	vmPolicy := config.VMControlPolicyConfig{
		IntervalSec: 1,   // Interval in seconds for the VM control commands
		TimeoutSec:  300, // Timeout in seconds for the VM control commands
	}
	vctx := vmcontroller.NewVMContext(vmInfo)
	vmController := vmcontroller.NewVMController(os.Stdin, os.Stdout, &vmControlConfig, &vmPolicy)

	err := vmController.StartVM(vctx)
	if err != nil {
		t.Errorf("StartVM failed: %v", err)
		return
	}
	if vctx.State() != vmstate.RUNNING {
		t.Errorf("Expected VM state to be OFFLINE, got %s", vctx.State())
		return
	}
	t.Logf("VM %s started successfully with IP %s", vctx.Machine(), vctx.IP())
	// sleep for a while to simulate VM startup time
	t.Logf("Waiting 10 sec...")
	time.Sleep(10 * time.Second)
	// stop the VM
	err = vmController.StopVM(vctx)
	if err != nil {
		t.Errorf("StopVM failed: %v", err)
		return
	}
	if vctx.State() != vmstate.STOPPED {
		t.Errorf("Expected VM state to be STOPPED, got %s", vctx.State())
		return
	}
	t.Logf("VM %s stopped successfully", vctx.Machine())
	// restore the snapshot
	err = vmController.RestoreSnapshot(vctx)
	if err != nil {
		t.Errorf("RestoreSnapshot failed: %v", err)
		return
	}
	if vctx.State() != vmstate.STOPPED {
		t.Errorf("Expected VM state to be STOPPED after restoring snapshot, got %s", vctx.State())
		return
	}
	t.Logf("Snapshot %s restored successfully for VM %s", vctx.Snapshot(), vctx.Machine())
}
