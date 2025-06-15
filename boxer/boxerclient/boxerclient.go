package boxer

import (
	"fmt"
	"os"

	"github.com/hongsam14/boxer/config"
	berror "github.com/hongsam14/boxer/error"
	"github.com/hongsam14/boxer/internal/vmcontroller"
)

// BoxerClient represents a request to perform an operation on a Box.
type BoxerClient interface {
	// Balloc allocates a Box for the given group.
	// It returns a Box instance or an error if allocation fails.
	// Check error code in github.com/hongsam14/boxer/error by using berror.Is(err, berror.Full)
	Balloc(group string) (Box, error)
	// Bfree frees the allocated Box.
	// It returns an error if the Box cannot be freed.
	Bfree(box Box) error
	// Do performs an operation on the Box.
	// The operation is specified in the BoxerRequest.
	// It returns a BoxerResponse with the result of the operation or an error if the operation fails.
	Do(req BoxerRequest) (BoxerResponse, error)
}

type boxerClient struct {
	config *config.BoxerConfig
	vmc    vmcontroller.VMController
	vc     vmcontroller.VMCompose
	// context pool key: group:machine, value: VMContext
	ctxPool map[string]*vmcontroller.VMContext
}

// NewBoxerClient creates a new BoxerClient with the provided configuration and file descriptors.
func NewBoxerClient(conf *config.BoxerConfig, fdin *os.File, fdout *os.File) (BoxerClient, error) {
	var err error

	// null check for configuration
	if conf == nil {
		return nil, berror.BoxerError{
			Code:   berror.InvalidArgument,
			Msg:    "error in NewBoxerClient",
			Origin: fmt.Errorf("configuration cannot be nil"),
		}
	}
	newClient := new(boxerClient)
	// dependency injection for configuration
	newClient.config = conf
	// Initialize context pool
	newClient.ctxPool = make(map[string]*vmcontroller.VMContext)
	// Initialize VMController and VMCompose with the provided configuration
	newClient.vmc = vmcontroller.NewVMController(
		fdin,
		fdout,
		&conf.VMControl,
		&conf.VMControlPolicy,
	)
	newClient.vc, err = vmcontroller.NewVMCompose(
		conf.VMInfo,
		&conf.VMControlPolicy,
	)
	return newClient, err
}

func (bc *boxerClient) generateContextPoolKey(group, machine string) string {
	// generate a key for the context pool
	return fmt.Sprintf("%s:%s", group, machine)
}

// Balloc allocates a Box for the given group.
// It returns a Box instance or an error if allocation fails.
// Check error code in github.com/hongsam14/boxer/error by using berror.Is(err, berror.Full)
func (bc *boxerClient) Balloc(group string) (Box, error) {
	// check validate the group parameter
	if group == "" {
		return nil, berror.BoxerError{
			Code:   berror.InvalidArgument,
			Msg:    "error in Balloc",
			Origin: fmt.Errorf("group cannot be empty"),
		}
	}
	vmCtx, err := bc.vc.AllocateVMContext(group)
	if err != nil {
		return nil, berror.BoxerError{
			Code:   berror.InternalError,
			Msg:    "error in Balloc",
			Origin: fmt.Errorf("failed to allocate Box: %w", err),
		}
	}
	// check if the VMContext is nil.
	if vmCtx == nil {
		// this case means VM allocation failed.
		// this can happen if all of the VMs in the group are already allocated
		// or maximum number of VMs in the group is reached.
		return nil, berror.BoxerError{
			Code:   berror.Full,
			Msg:    "error in Balloc",
			Origin: fmt.Errorf("no available VM to be allocated in this env"),
		}
	}
	// check if the VMContext exists in the context pool
	key := bc.generateContextPoolKey(vmCtx.Group(), vmCtx.Machine())
	if _, exists := bc.ctxPool[key]; exists {
		return nil, berror.BoxerError{
			Code:   berror.InternalError,
			Msg:    "error in Balloc",
			Origin: fmt.Errorf("VMContext already exists in the context pool for group %s and machine %s", vmCtx.Group(), vmCtx.Machine()),
		}
	}
	// store the VMContext in the context pool
	bc.ctxPool[key] = vmCtx
	// create a new Box instance with the VMContext
	return NewBox(vmCtx), nil
}

// Bfree frees the allocated Box.
// It returns an error if the Box cannot be freed.
func (bc *boxerClient) Bfree(box Box) error {
	// check if the box parameter is nil
	if box == nil {
		return berror.BoxerError{
			Code:   berror.InvalidArgument,
			Msg:    "error in Bfree",
			Origin: fmt.Errorf("box cannot be nil"),
		}
	}
	key := bc.generateContextPoolKey(box.Group(), box.Machine())
	if _, exists := bc.ctxPool[key]; !exists {
		return berror.BoxerError{
			Code:   berror.InternalError,
			Msg:    "error in Bfree",
			Origin: fmt.Errorf("VMContext does not exist in the context pool for group %s and machine %s", box.Group(), box.Machine()),
		}
	}
	// free the VMContext using the VMController
	if err := bc.vc.FreeVMContext(bc.ctxPool[key]); err != nil {
		return berror.BoxerError{
			Code:   berror.InternalError,
			Msg:    "error in Bfree",
			Origin: fmt.Errorf("failed to free Box: %w", err),
		}
	}
	// delete the VMContext from the context pool
	delete(bc.ctxPool, key)
	return nil
}

// Do performs an operation on the Box.
// The operation is specified in the BoxerRequest.
// It returns a BoxerResponse with the result of the operation or an error if the operation fails.
func (bc *boxerClient) Do(req BoxerRequest) (BoxerResponse, error) {
	var err error

	// check if the request is valid
	if req.BoxInfo == nil {
		return BoxerResponse{
				Code:    INVALID_REQUEST,
				BoxInfo: req.BoxInfo,
			},
			berror.BoxerError{
				Code:   berror.InvalidArgument,
				Msg:    "error in Do",
				Origin: fmt.Errorf("box info cannot be nil"),
			}
	}
	// check if box is allocated
	key := bc.generateContextPoolKey(req.BoxInfo.Group(), req.BoxInfo.Machine())
	vmCtx, exists := bc.ctxPool[key]
	if !exists {
		return BoxerResponse{
				Code:    NOT_FOUND,
				BoxInfo: req.BoxInfo,
			},
			berror.BoxerError{
				Code:   berror.InvalidState,
				Msg:    "error in Do",
				Origin: fmt.Errorf("box is not allocated for group %s and machine %s", req.BoxInfo.Group(), req.BoxInfo.Machine()),
			}
	}
	// operate on the VMContext based on the request operation
	switch req.OP {
	case STOP:
		// stop the VM
		err = bc.vmc.StopVM(vmCtx)
	case START:
		// start the VM
		err = bc.vmc.StartVM(vmCtx)
	case RESTORE:
		// restore the VM from a snapshot
		err = bc.vmc.RestoreSnapshot(vmCtx)
	}
	if err != nil {
		return BoxerResponse{
				Code:    INTERNAL_ERROR,
				BoxInfo: NewBox(vmCtx),
			},
			berror.BoxerError{
				Code:   berror.InternalError,
				Msg:    "error in Do",
				Origin: fmt.Errorf("failed to perform operation %s on Box: %w", req.OP, err),
			}
	}
	return BoxerResponse{
		Code:    SUCCESS,
		BoxInfo: NewBox(vmCtx),
	}, nil
}
