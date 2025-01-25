package exec

import (
	"sync"
)

type Promise[t any] struct {
	wait  sync.WaitGroup
	value t
}
