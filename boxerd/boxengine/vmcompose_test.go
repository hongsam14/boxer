package boxengine_test

import (
	"boxerd/boxengine"
	"boxerd/config"
	"boxerd/vmcontroller"
	"math/rand"
	"testing"
)

func TestVMComposeAllocateAndFree(t *testing.T) {
	vmInfoMap := map[string]config.VMInfoConfig{
		"vm1": {
			Name:     "vm1",
			Snapshot: "snapshot1",
			IP:       "127.0.0.1",
			OS:       "linux",
			Group:    "group1",
		},
		"vm2": {
			Name:     "vm2",
			Snapshot: "snapshot2",
			IP:       "127.0.0.2",
			OS:       "linux",
			Group:    "group1",
		},
		"vm3": {
			Name:     "vm3",
			Snapshot: "snapshot3",
			IP:       "127.0.0.1",
			OS:       "linux",
			Group:    "group2",
		},
	}
	vmPolicy := config.VMControlPolicyConfig{
		IntervalSec:     10,
		TimeoutSec:      30,
		MaxVMOperations: 3,
	}
	vmCompose, err := boxengine.NewVMCompose(vmInfoMap, &vmPolicy)
	if err != nil {
		t.Fatalf("Failed to create VMCompose: %v", err)
		return
	}
	// Allocate VMs
	vm1, err := vmCompose.AllocateVMContext("group1")
	if err != nil {
		t.Fatalf("Failed to allocate VM1: %v", err)
		return
	}
	if vm1 == nil {
		t.Fatal("Allocated VM1 is nil")
		return
	}
	t.Logf("Allocated VM1: %s %s", vm1.Machine(), vm1.Group())
	// Allocate another VM in the same group
	vm2, err := vmCompose.AllocateVMContext("group1")
	if err != nil {
		t.Fatalf("Failed to allocate VM2: %v", err)
		return
	}
	if vm2 == nil {
		t.Fatal("Allocated VM2 is nil")
		return
	}
	t.Logf("Allocated VM2: %s %s", vm2.Machine(), vm2.Group())
	// Allocate a VM in a different group
	vm3, err := vmCompose.AllocateVMContext("group2")
	if err != nil {
		t.Fatalf("Failed to allocate VM3: %v", err)
		return
	}
	if vm3 == nil {
		t.Fatal("Allocated VM3 is nil")
		return
	}
	t.Logf("Allocated VM3: %s %s", vm3.Machine(), vm3.Group())
	// Free the allocated VMs
	err = vmCompose.FreeVMContext(vm1)
	if err != nil {
		t.Fatalf("Failed to free VM1: %v", err)
		return
	}
	err = vmCompose.FreeVMContext(vm2)
	if err != nil {
		t.Fatalf("Failed to free VM2: %v", err)
		return
	}
	err = vmCompose.FreeVMContext(vm3)
	if err != nil {
		t.Fatalf("Failed to free VM3: %v", err)
		return
	}
	t.Log("Successfully freed all allocated VMs")
	// Check if all VMs are freed by trying to allocate again
	_, err = vmCompose.AllocateVMContext("group1")
	if err != nil {
		t.Fatal("Expected error when allocating VM after freeing, but got none")
		return
	}
	t.Log("Successfully allocated a new VM after freeing all previous VMs")
	_, err = vmCompose.AllocateVMContext("group1")
	if err != nil {
		t.Fatal("Expected error when allocating VM after freeing, but got none")
		return
	}
	t.Log("Successfully allocated another new VM after freeing all previous VMs")
	_, err = vmCompose.AllocateVMContext("group2")
	if err != nil {
		t.Fatal("Expected error when allocating VM after freeing, but got none")
		return
	}
	t.Log("Successfully allocated a new VM in a different group after freeing all previous VMs")
	// Final check to ensure all VMs are allocated correctly
	out, err := vmCompose.AllocateVMContext("group1")
	if out != nil {
		t.Fatalf("Expected no VM to be allocated after freeing, but got: %s", out.Machine())
		return
	}
}

func TestVMComposeAllocateNonExistentGroup(t *testing.T) {
	vmInfoMap := map[string]config.VMInfoConfig{
		"vm1": {
			Name:     "vm1",
			Snapshot: "snapshot1",
			IP:       "127.0.0.1",
			OS:       "linux",
			Group:    "group1",
		},
		"vm2": {
			Name:     "vm2",
			Snapshot: "snapshot2",
			IP:       "127.0.0.2",
			OS:       "linux",
			Group:    "group1",
		},
		"vm3": {
			Name:     "vm3",
			Snapshot: "snapshot3",
			IP:       "127.0.0.1",
			OS:       "linux",
			Group:    "group2",
		},
	}
	vmPolicy := config.VMControlPolicyConfig{
		IntervalSec:     10,
		TimeoutSec:      30,
		MaxVMOperations: 3,
	}
	vmCompose, err := boxengine.NewVMCompose(vmInfoMap, &vmPolicy)
	if err != nil {
		t.Fatalf("Failed to create VMCompose: %v", err)
		return
	}
	// Try to allocate a VM from a non-existent group
	out, err := vmCompose.AllocateVMContext("nonExistentGroup")
	if err == nil || out != nil {
		t.Fatal("Expected error when allocating VM from non-existent group, but got none")
		return
	}
	t.Logf("Successfully caught error when allocating VM from non-existent group: %v", err)
}

func TestVMComposeAllocateExceedingLimit(t *testing.T) {
	vmInfoMap := map[string]config.VMInfoConfig{
		"vm1": {
			Name:     "vm1",
			Snapshot: "snapshot1",
			IP:       "127.0.0.1",
			OS:       "linux",
			Group:    "group1",
		},
		"vm2": {
			Name:     "vm2",
			Snapshot: "snapshot2",
			IP:       "127.0.0.2",
			OS:       "linux",
			Group:    "group1",
		},
		"vm3": {
			Name:     "vm3",
			Snapshot: "snapshot3",
			IP:       "127.0.0.1",
			OS:       "linux",
			Group:    "group2",
		},
	}
	vmPolicy := config.VMControlPolicyConfig{
		IntervalSec:     10,
		TimeoutSec:      30,
		MaxVMOperations: 2,
	}
	vmCompose, err := boxengine.NewVMCompose(vmInfoMap, &vmPolicy)
	if err != nil {
		t.Fatalf("Failed to create VMCompose: %v", err)
		return
	}
	// Allocate VMs up to the limit
	vm1, err := vmCompose.AllocateVMContext("group1")
	if err != nil {
		t.Fatalf("Failed to allocate VM1: %v", err)
		return
	}
	if vm1 == nil {
		t.Fatal("Allocated VM1 is nil")
		return
	}
	vm2, err := vmCompose.AllocateVMContext("group2")
	if err != nil {
		t.Fatalf("Failed to allocate VM2: %v", err)
		return
	}
	if vm2 == nil {
		t.Fatal("Allocated VM2 is nil")
		return
	}
	// Try to allocate a third VM which should exceed the limit
	out, err := vmCompose.AllocateVMContext("group1")
	if err != nil {
		t.Fatal("error when allocating VM exceeding limit")
		return
	}
	if out != nil {
		t.Fatalf("Expected no VM to be allocated when exceeding limit, but got: %s", out.Machine())
		return
	}
	t.Logf("Successfully return nil when allocating VM exceeding limit: %v", out)
}

func TestVMComposeAllocateMaximumNumberOfVMsGroup(t *testing.T) {
	vmInfoMap := map[string]config.VMInfoConfig{
		"vm1": {
			Name:     "vm1",
			Snapshot: "snapshot1",
			IP:       "127.0.0.1",
			OS:       "linux",
			Group:    "group1",
		},
		"vm2": {
			Name:     "vm2",
			Snapshot: "snapshot2",
			IP:       "127.0.0.2",
			OS:       "linux",
			Group:    "group1",
		},
		"vm3": {
			Name:     "vm3",
			Snapshot: "snapshot3",
			IP:       "127.0.0.1",
			OS:       "linux",
			Group:    "group2",
		},
	}
	vmPolicy := config.VMControlPolicyConfig{
		IntervalSec:     10,
		TimeoutSec:      30,
		MaxVMOperations: 3,
	}
	vmCompose, err := boxengine.NewVMCompose(vmInfoMap, &vmPolicy)
	if err != nil {
		t.Fatalf("Failed to create VMCompose: %v", err)
		return
	}
	// Allocate VMs up to the limit
	vm1, err := vmCompose.AllocateVMContext("group1")
	if err != nil {
		t.Fatalf("Failed to allocate VM1: %v", err)
		return
	}
	if vm1 == nil {
		t.Fatal("Allocated VM1 is nil")
		return
	}
	vm2, err := vmCompose.AllocateVMContext("group1")
	if err != nil {
		t.Fatalf("Failed to allocate VM2: %v", err)
		return
	}
	if vm2 == nil {
		t.Fatal("Allocated VM2 is nil")
		return
	}
	// Try to allocate a third VM which should exceed the limit
	out, err := vmCompose.AllocateVMContext("group1")
	if err != nil {
		t.Fatal("error when allocating VM exceeding limit")
		return
	}
	if out != nil {
		t.Fatalf("Expected no VM to be allocated when exceeding limit, but got: %s", out.Machine())
		return
	}
	t.Logf("Successfully returned nil when allocating VM exceeding limit: %v", out)
}

func TestVMComposeFreeDuplicatedVMContext(t *testing.T) {
	vmInfoMap := map[string]config.VMInfoConfig{
		"vm1": {
			Name:     "vm1",
			Snapshot: "snapshot1",
			IP:       "127.0.0.1",
			OS:       "linux",
			Group:    "group1",
		},
		"vm2": {
			Name:     "vm2",
			Snapshot: "snapshot2",
			IP:       "127.0.0.2",
			OS:       "linux",
			Group:    "group1",
		},
		"vm3": {
			Name:     "vm3",
			Snapshot: "snapshot3",
			IP:       "127.0.0.1",
			OS:       "linux",
			Group:    "group2",
		},
	}
	vmPolicy := config.VMControlPolicyConfig{
		IntervalSec:     10,
		TimeoutSec:      30,
		MaxVMOperations: 2,
	}
	vmCompose, err := boxengine.NewVMCompose(vmInfoMap, &vmPolicy)
	if err != nil {
		t.Fatalf("Failed to create VMCompose: %v", err)
		return
	}
	// Allocate VMs up to the limit
	vm1, err := vmCompose.AllocateVMContext("group1")
	if err != nil {
		t.Fatalf("Failed to allocate VM1: %v", err)
		return
	}
	if vm1 == nil {
		t.Fatal("Allocated VM1 is nil")
		return
	}
	// Free the VM once
	err = vmCompose.FreeVMContext(vm1)
	if err != nil {
		t.Fatalf("Failed to free VM1: %v", err)
		return
	}
	// Try to free the same VM again
	err = vmCompose.FreeVMContext(vm1)
	if err == nil {
		t.Fatal("Expected error when freeing a VM that is already freed, but got none")
		return
	}
	t.Logf("Successfully caught error when freeing a VM that is already freed: %v", err)
}

func TestVMComposeFreeNonExistentVMContext(t *testing.T) {
	vmInfoMap := map[string]config.VMInfoConfig{
		"vm1": {
			Name:     "vm1",
			Snapshot: "snapshot1",
			IP:       "127.0.0.1",
			OS:       "linux",
			Group:    "group1",
		},
		"vm2": {
			Name:     "vm2",
			Snapshot: "snapshot2",
			IP:       "127.0.0.2",
			OS:       "linux",
			Group:    "group1",
		},
		"vm3": {
			Name:     "vm3",
			Snapshot: "snapshot3",
			IP:       "127.0.0.1",
			OS:       "linux",
			Group:    "group2",
		},
	}
	vmPolicy := config.VMControlPolicyConfig{
		IntervalSec:     10,
		TimeoutSec:      30,
		MaxVMOperations: 2,
	}
	vmCompose, err := boxengine.NewVMCompose(vmInfoMap, &vmPolicy)
	if err != nil {
		t.Fatalf("Failed to create VMCompose: %v", err)
		return
	}
	// Try to free a VM that does not exist
	nonExistentVM := vmcontroller.NewVMContext(config.VMInfoConfig{
		Name:     "nonExistentVM",
		Snapshot: "nonExistentSnapshot",
		IP:       "127.0.0.1",
		OS:       "linux",
		Group:    "nonExistentGroup",
	})
	err = vmCompose.FreeVMContext(nonExistentVM)
	if err == nil {
		t.Fatal("Expected error when freeing a non-existent VM, but got none")
		return
	}
	t.Logf("Successfully caught error when freeing a non-existent VM: %v", err)
}

func TestVMComposeAllocateAndFreeMultipleLoops(t *testing.T) {
	vmInfoMap := map[string]config.VMInfoConfig{
		"vm1": {
			Name:     "vm1",
			Snapshot: "snapshot1",
			IP:       "127.0.0.1",
			OS:       "linux",
			Group:    "group1",
		},
		"vm2": {
			Name:     "vm2",
			Snapshot: "snapshot2",
			IP:       "127.0.0.2",
			OS:       "linux",
			Group:    "group1",
		},
		"vm3": {
			Name:     "vm3",
			Snapshot: "snapshot3",
			IP:       "127.0.0.1",
			OS:       "linux",
			Group:    "group2",
		},
	}
	vmPolicy := config.VMControlPolicyConfig{
		IntervalSec:     10,
		TimeoutSec:      30,
		MaxVMOperations: 2,
	}
	vmCompose, err := boxengine.NewVMCompose(vmInfoMap, &vmPolicy)
	if err != nil {
		t.Fatalf("Failed to create VMCompose: %v", err)
		return
	}
	// Loop to allocate and free VMs random multiple times
	for i := 0; i < rand.Intn(10)+1; i++ {
		t.Logf("Loop iteration %d", i+1)
		// Allocate VMs
		vm1, err := vmCompose.AllocateVMContext("group1")
		if err != nil {
			t.Fatalf("Failed to allocate VM1: %v", err)
			return
		}
		if vm1 == nil {
			t.Fatal("Allocated VM1 is nil")
			return
		}
		t.Logf("Allocated VM1: %s %s", vm1.Machine(), vm1.Group())

		vm2, err := vmCompose.AllocateVMContext("group2")
		if err != nil {
			t.Fatalf("Failed to allocate VM2: %v", err)
			return
		}
		if vm2 == nil {
			t.Fatal("Allocated VM2 is nil")
			return
		}
		t.Logf("Allocated VM2: %s %s", vm2.Machine(), vm2.Group())

		// Free the allocated VMs
		err = vmCompose.FreeVMContext(vm1)
		if err != nil {
			t.Fatalf("Failed to free VM1: %v", err)
			return
		}

		err = vmCompose.FreeVMContext(vm2)
		if err != nil {
			t.Fatalf("Failed to free VM2: %v", err)
			return
		}

		t.Logf("Successfully freed VM1 and VM2 in iteration %d", i+1)
	}

}
