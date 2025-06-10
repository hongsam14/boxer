package boxer

import (
	"boxerd/internal/vmcontroller"
)

type boxerClient struct {
	vmc vmcontroller.VMController
	vc  vmcontroller.VMCompose
}

type boxerRequest struct {
}

type boxerResponse struct {
}
