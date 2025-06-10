package vmcontroller

import (
	"boxerd/config"
	berror "boxerd/error"
	"fmt"
	"sync/atomic"
)

// VMCompose allocates and Frees VMContexts based on the VMInfoMap and VMPolicy.
// It manages groups of VMContexts and ensures that the maximum number of VM operations is not exceeded.
// It provides methods to allocate and free VMContexts from the specified groups.
type VMCompose interface {
	// AllocateVMContext allocates a VMContext from the specified group.
	// It returns the VMContext if available, or nil if all VMContexts are allocated.
	AllocateVMContext(groupName string) (*VMContext, error)
	// FreeVMContext frees a VMContext and adds it back to the group.
	FreeVMContext(free *VMContext) error
}

type vmCompose struct {
	groupMap            map[string]*vmContextGroup
	maxVMOperations     uint32
	currentVMOperations uint32
}

// NewVMCompose creates a new vmCompose with the given VMInfoMap and VMPolicy.
func NewVMCompose(vmInfoMap map[string]config.VMInfoConfig, vmPolicy *config.VMControlPolicyConfig) (VMCompose, error) {
	if vmPolicy == nil {
		return nil, berror.BoxerError{
			Code:   berror.InvalidArgument,
			Msg:    "error in boxCompose NewVMCompose",
			Origin: fmt.Errorf("vmPolicy cannot be nil"),
		}
	}

	newCompose := new(vmCompose)
	newCompose.groupMap = make(map[string]*vmContextGroup)
	newCompose.maxVMOperations = uint32(vmPolicy.MaxVMOperations)
	newCompose.currentVMOperations = 0

	// create groupMap based on the VMInfoMap
	for _, vmInfo := range vmInfoMap {
		if group, exists := newCompose.groupMap[vmInfo.Group]; exists {
			// If the group already exists, append the VMContext to the existing group
			err := group.AppendVMContext(vmInfo)
			if err != nil {
				return nil, berror.BoxerError{
					Code:   berror.InvalidOperation,
					Msg:    "error in boxCompose NewVMCompose",
					Origin: fmt.Errorf("failed to append VMContext %s to group %s: %w", vmInfo.Name, vmInfo.Group, err),
				}
			}
		} else {
			// If the group does not exist, create a new group with the VMContext
			newGroup, err := newVMGroup(vmInfo.Group, vmInfo)
			if err != nil {
				return nil, berror.BoxerError{
					Code:   berror.InvalidOperation,
					Msg:    "error in boxCompose NewVMCompose",
					Origin: fmt.Errorf("failed to create new VMContextGroup for group %s: %w", vmInfo.Group, err),
				}
			}
			newCompose.groupMap[vmInfo.Group] = newGroup
		}
	}

	return newCompose, nil
}

// AllocateVMContext allocates a VMContext from the specified group.
// It returns the VMContext if available, or nil if all VMContexts are allocated.
// It also checks if the group exists and if the maximum number of VM operations has been reached.
// If the maximum number of VM operations is reached, it returns nil without an error.
func (vc *vmCompose) AllocateVMContext(groupName string) (*VMContext, error) {
	// check if the group exists
	group, exists := vc.groupMap[groupName]
	if !exists {
		return nil, berror.BoxerError{
			Code:   berror.InvalidArgument,
			Msg:    "error in boxCompose AllocateVMContext",
			Origin: fmt.Errorf("group %s does not exist", groupName),
		}
	}
	// check if the group has reached the maximum number of VM operations
	if atomic.LoadUint32(&vc.currentVMOperations) >= vc.maxVMOperations {
		return nil, nil // no VMContext can be allocated because the maximum number of VM operations is reached
	}
	// allocate a VMContext from the group
	vmContext, err := group.AllocateVMContext()
	if err != nil {
		return nil, berror.BoxerError{
			Code:   berror.InvalidOperation,
			Msg:    "error in boxCompose AllocateVMContext",
			Origin: fmt.Errorf("failed to allocate VMContext from group %s: %w", groupName, err),
		}
	}
	if vmContext == nil {
		// if the vmContext is nil, it means that all VMContexts are allocated
		return nil, nil
	}
	// increment the current VM operations count
	// compare and swap the currentVMOperations with the incremented value
	atomic.AddUint32(&vc.currentVMOperations, 1)
	return vmContext, nil
}

// FreeVMContext frees a VMContext and adds it back to the group.
// It checks if the group exists and if the VMContext is in the allocated VMContexts.
// It decrements the current VM operations count after freeing the VMContext.
// If the group does not exist or the VMContext is not allocated, it returns an error.
func (vc *vmCompose) FreeVMContext(free *VMContext) error {
	// check if the group exists
	group, exists := vc.groupMap[free.Group()]
	if !exists {
		return berror.BoxerError{
			Code:   berror.InvalidArgument,
			Msg:    "error in boxCompose FreeVMContext",
			Origin: fmt.Errorf("group %s does not exist", free.Group()),
		}
	}
	// free the VMContext in the group
	err := group.FreeVMContext(free)
	if err != nil {
		return berror.BoxerError{
			Code:   berror.InvalidOperation,
			Msg:    "error in boxCompose FreeVMContext",
			Origin: fmt.Errorf("failed to free VMContext %s in group %s: %w", free.Machine(), free.Group(), err),
		}
	}
	// decrement the current VM operations count
	atomic.AddUint32(&vc.currentVMOperations, ^uint32(0)) // decrement by 1
	return nil
}

// vmContextGroup is a struct that holds a group of VMContexts.
// It is used to manage the allocation and deallocation of VMContexts.
type vmContextGroup struct {
	groupName       string
	size            int
	vmInfoPool      []*VMContext
	allocatedVMInfo map[string]*VMContext
}

// GroupName returns the name of the VMContextGroup.
func (vg *vmContextGroup) GroupName() string {
	return vg.groupName
}

// newVMGroup creates a new vmContextGroup with the given groupName and VMInfos.
func newVMGroup(groupName string, vmInfos ...config.VMInfoConfig) (*vmContextGroup, error) {
	newGroup := new(vmContextGroup)
	newGroup.groupName = groupName
	if groupName == "" {
		return nil, berror.BoxerError{
			Code:   berror.InvalidArgument,
			Msg:    "error in boxGroup newBoxGroup",
			Origin: fmt.Errorf("group name cannot be empty"),
		}
	}
	// allocate the vmInfoPool with the given vmInfos
	newGroup.vmInfoPool = make([]*VMContext, len(vmInfos))
	for idx, vmInfo := range vmInfos {
		if groupName != vmInfo.Group {
			return nil, berror.BoxerError{
				Code:   berror.InvalidArgument,
				Msg:    "error in boxGroup newBoxGroup",
				Origin: fmt.Errorf("VM %s does not belong to group %s", vmInfo.Name, groupName),
			}
		}
		newGroup.vmInfoPool[idx] = NewVMContext(vmInfo)
	}
	if len(newGroup.vmInfoPool) == 0 {
		return nil, berror.BoxerError{
			Code:   berror.InvalidOperation,
			Msg:    "error in boxGroup newBoxGroup",
			Origin: fmt.Errorf("group %s has no VM info", groupName),
		}
	}
	// allocate the allocatedVMInfo map
	newGroup.allocatedVMInfo = make(map[string]*VMContext)
	// set the size of the group to the number of VM infos
	newGroup.size = len(vmInfos)
	return newGroup, nil
}

// AllocateVMContext allocates a VMContext from the group.
// It returns the VMContext if available, or nil if all VMContexts are allocated.
func (vg *vmContextGroup) AllocateVMContext() (*VMContext, error) {
	// get the first VMContext from the pool
	if len(vg.vmInfoPool) == 0 {
		// all the VMContexts are allocated, return an nil because there is no VMContext available.
		// and this is not an error, just a normal case.
		return nil, nil
	}
	vmContext := vg.vmInfoPool[0]
	// remove the VMContext from the pool
	vg.vmInfoPool = vg.vmInfoPool[1:]
	// add the VMContext to the allocatedVMInfo map
	vg.allocatedVMInfo[vmContext.Machine()] = vmContext
	// check if the size of the group is equal to the size of the vmInfoPool + len(allocatedVMInfo)
	if len(vg.vmInfoPool)+len(vg.allocatedVMInfo) != vg.size {
		// return an error if the size is not equal
		return nil, berror.BoxerError{
			Code:   berror.InvalidOperation,
			Msg:    "error in boxGroup AllocateVMContext",
			Origin: fmt.Errorf("group %s size mismatch: expected %d, got %d", vg.groupName, vg.size, len(vg.vmInfoPool)+len(vg.allocatedVMInfo)),
		}
	}
	// return the VMContext
	return vmContext, nil
}

// FreeVMContext frees a VMContext and adds it back to the pool.
func (bg *vmContextGroup) FreeVMContext(vmContext *VMContext) error {
	// check if the VMContext is in the allocatedVMInfo map
	if _, exists := bg.allocatedVMInfo[vmContext.Machine()]; !exists {
		return berror.BoxerError{
			Code:   berror.InvalidArgument,
			Msg:    "error in boxGroup FreeVMContext",
			Origin: fmt.Errorf("VMContext %s is not allocated in group %s", vmContext.Machine(), bg.groupName),
		}
	}
	// remove the VMContext from the allocatedVMInfo map
	delete(bg.allocatedVMInfo, vmContext.Machine())
	// add the VMContext back to the vmInfoPool
	bg.vmInfoPool = append(bg.vmInfoPool, vmContext)
	// check if the size of the group is equal to the size of the vmInfoPool + len(allocatedVMInfo)
	if len(bg.vmInfoPool)+len(bg.allocatedVMInfo) != bg.size {
		return berror.BoxerError{
			Code:   berror.InvalidOperation,
			Msg:    "error in boxGroup FreeVMContext",
			Origin: fmt.Errorf("group %s size mismatch: expected %d, got %d", bg.groupName, bg.size, len(bg.vmInfoPool)+len(bg.allocatedVMInfo)),
		}
	}
	return nil
}

// AppendVMContext appends a VMContext to the group.
func (bg *vmContextGroup) AppendVMContext(added config.VMInfoConfig) error {
	// check if the VMContext is already in the allocatedVMInfo map
	if _, exists := bg.allocatedVMInfo[added.Name]; exists {
		return berror.BoxerError{
			Code:   berror.InvalidArgument,
			Msg:    "error in boxGroup AppendVMContext",
			Origin: fmt.Errorf("VMContext %s is already allocated in group %s", added.Name, bg.groupName),
		}
	}
	// check if the VMContext is already in the vmInfoPool
	for _, existingVMContext := range bg.vmInfoPool {
		if existingVMContext.Machine() == added.Name {
			return berror.BoxerError{
				Code:   berror.InvalidArgument,
				Msg:    "error in boxGroup AppendVMContext",
				Origin: fmt.Errorf("VMContext %s is already in the vmInfoPool of group %s", added.Name, bg.groupName),
			}
		}
	}
	// add the VMContext to the vmInfoPool
	addedVMContext := NewVMContext(added)
	bg.vmInfoPool = append(bg.vmInfoPool, addedVMContext)
	// increase the size of the group
	bg.size++
	// check if the size of the group is equal to the size of the vmInfoPool + len(allocatedVMInfo)
	if len(bg.vmInfoPool)+len(bg.allocatedVMInfo) != bg.size {
		return berror.BoxerError{
			Code:   berror.InvalidOperation,
			Msg:    "error in boxGroup AppendVMContext",
			Origin: fmt.Errorf("group %s size mismatch: expected %d, got %d", bg.groupName, bg.size, len(bg.vmInfoPool)+len(bg.allocatedVMInfo)),
		}
	}
	return nil
}
