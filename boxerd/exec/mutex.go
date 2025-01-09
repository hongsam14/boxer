package exec

import (
	"boxerd/error"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Cushion struct {
	releaseMux sync.Mutex
	IsLocked   int32
	IsWaiting  int32
	// timer
	period uint
}

func InitCushion(period uint) *Cushion {
	newCushion := new(Cushion)
	newCushion.releaseMux = sync.Mutex{}
	atomic.StoreInt32(&newCushion.IsLocked, 0)
	atomic.StoreInt32(&newCushion.IsWaiting, 0)
	return newCushion
}

func (cushion *Cushion) Lock() {
	// lock the releaseMux
	cushion.releaseMux.Lock()
	atomic.StoreInt32(&cushion.IsLocked, 1)
}

func (cushion *Cushion) Release() {
	if atomic.LoadInt32(&cushion.IsLocked) == 0 {
		panic(error.BoxerError{
			Code:   error.InvalidOperation,
			Msg:    "error while cushion.release()",
			Origin: fmt.Errorf("cushion is released before locked"),
		})
	}
	if atomic.LoadInt32(&cushion.IsWaiting) == 1 {
		panic(error.BoxerError{
			Code:   error.InvalidOperation,
			Msg:    "error while cushion.release()",
			Origin: fmt.Errorf("cushion is released while waiting"),
		})
	}
	atomic.StoreInt32(&cushion.IsWaiting, 1)
	go cushion.timerThread()
}

func (cuchion *Cushion) timerThread() {
	// wait for the period
	time.Sleep(time.Duration(cuchion.period) * time.Second)

	if atomic.LoadInt32(&cuchion.IsLocked) == 1 {
		cuchion.releaseMux.Unlock()
	}
	atomic.StoreInt32(&cuchion.IsWaiting, 0)
	atomic.StoreInt32(&cuchion.IsLocked, 0)
}
