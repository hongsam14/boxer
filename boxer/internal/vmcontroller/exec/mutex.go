package exec

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hongsam14/boxer/error"
)

type PaddedMutex struct {
	releaseMux sync.Mutex
	IsLocked   int32
	IsWaiting  int32
	// timer
	period uint
}

func InitPaddedMutex(period uint) *PaddedMutex {
	newPaddedMux := new(PaddedMutex)
	newPaddedMux.period = period
	newPaddedMux.releaseMux = sync.Mutex{}
	atomic.StoreInt32(&newPaddedMux.IsLocked, 0)
	atomic.StoreInt32(&newPaddedMux.IsWaiting, 0)
	return newPaddedMux
}

func (pMux *PaddedMutex) Lock() {
	// lock the releaseMux
	pMux.releaseMux.Lock()
	atomic.StoreInt32(&pMux.IsLocked, 1)
}

func (pMux *PaddedMutex) Release() {
	if atomic.LoadInt32(&pMux.IsLocked) == 0 {
		panic(error.BoxerError{
			Code:   error.InvalidOperation,
			Msg:    "error while pMux.release()",
			Origin: fmt.Errorf("pMux is released before locked"),
		})
	}
	if atomic.LoadInt32(&pMux.IsWaiting) == 1 {
		panic(error.BoxerError{
			Code:   error.InvalidOperation,
			Msg:    "error while pMux.release()",
			Origin: fmt.Errorf("pMux is released while waiting"),
		})
	}
	atomic.StoreInt32(&pMux.IsWaiting, 1)
	go pMux.timerThread()
}

func (cuchion *PaddedMutex) timerThread() {
	// wait for the period
	time.Sleep(time.Duration(cuchion.period) * time.Second)

	if atomic.LoadInt32(&cuchion.IsLocked) == 1 {
		cuchion.releaseMux.Unlock()
	}
	atomic.StoreInt32(&cuchion.IsWaiting, 0)
	atomic.StoreInt32(&cuchion.IsLocked, 0)
}
