package boxer_test

import (
	"os"
	"testing"
	"time"

	boxer "github.com/hongsam14/boxer/boxerclient"
	"github.com/hongsam14/boxer/config"
)

var testConfig = &config.BoxerConfig{
	VMInfo: map[string]config.VMInfoConfig{
		"sb_win10_develop_v2": {
			Name:     "sb_win10_develop_v2",
			Snapshot: "snapshot0",
			OS:       "windows",
			Group:    "testGroup",
			IP:       "127.0.0.1",
		},
		"sb_win10_develop_v2_clone_0": {
			Name:     "sb_win10_develop_v2_clone_0",
			Snapshot: "snapshot0",
			OS:       "windows",
			Group:    "testGroup",
			IP:       "127.0.0.2",
		},
		"openssh": {
			Name:     "openssh",
			Snapshot: "Snapshot 1",
			OS:       "linux",
			Group:    "testGroup2",
			IP:       "127.0.0.3",
		},
	},
	VMControl: config.VMControlConfig{
		StartCmd:           "VBoxManage startvm $machine",
		StopCmd:            "VBoxManage controlvm $machine poweroff",
		RestoreSnapshotCmd: "VBoxManage snapshot $machine restore $snapshot",
	},
	VMControlPolicy: config.VMControlPolicyConfig{
		IntervalSec:     1,
		TimeoutSec:      30,
		MaxVMOperations: 3,
	},
}

func TestNewBoxerClient(t *testing.T) {
	client, err := boxer.NewBoxerClient(testConfig, os.Stdin, os.Stdout)
	if err != nil {
		t.Fatalf("Failed to create BoxerClient: %v", err)
		return
	}

	if client == nil {
		t.Fatal("BoxerClient is nil")
		return
	}

	// allocate a Box for the test group
	box, err := client.Balloc("testGroup2")
	if err != nil {
		t.Fatalf("Failed to allocate Box: %v", err)
		return
	}

	// start the Box
	resp, err := client.Do(boxer.BoxerRequest{
		BoxInfo: box,
		OP:      boxer.START,
	})
	if err != nil {
		t.Fatalf("Failed to start Box: %v", err)
		return
	}
	if resp.Code != boxer.SUCCESS {
		t.Fatalf("Expected SUCCESS, got %s", resp.Code)
		return
	}
	t.Logf("Box started successfully: %s %v", resp.BoxInfo.Machine(), resp.BoxInfo.State().String())
	// sleep for a while to simulate VM startup time
	t.Logf("Waiting 20 sec...")
	time.Sleep(20 * time.Second)

	// stop the Box
	resp, err = client.Do(boxer.BoxerRequest{
		BoxInfo: box,
		OP:      boxer.STOP,
	})
	if err != nil {
		t.Fatalf("Failed to stop Box: %v", err)
		return
	}
	if resp.Code != boxer.SUCCESS {
		t.Fatalf("Expected SUCCESS, got %s", resp.Code)
		return
	}
	t.Logf("Box stopped successfully: %s %v", resp.BoxInfo.Machine(), resp.BoxInfo.State().String())

	// restore the snapshot
	resp, err = client.Do(boxer.BoxerRequest{
		BoxInfo: box,
		OP:      boxer.RESTORE,
	})
	if err != nil {
		t.Fatalf("Failed to restore Box: %v", err)
		return
	}
	if resp.Code != boxer.SUCCESS {
		t.Fatalf("Expected SUCCESS, got %s", resp.Code)
		return
	}
	t.Logf("Box restored successfully: %s %v", resp.BoxInfo.Machine(), resp.BoxInfo.State().String())
	// deallocate the Box
	err = client.Bfree(box)
	if err != nil {
		t.Fatalf("Failed to deallocate Box: %v", err)
		return
	}
	t.Logf("Box deallocated successfully: %s", box.Machine())
}
