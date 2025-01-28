package exec

import (
	berror "boxerd/error"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

type Promise struct {
	ctx     context.Context
	cancel  context.CancelFunc
	cmd     exec.Cmd
	eg      *errgroup.Group
	waitCnt int32
}

func Run(inStream *os.File, outStream *os.File, arg0 string, args ...string) (promise *Promise) {
	promise = new(Promise)
	promise.ctx, promise.cancel = context.WithCancel(context.Background())
	// set conditional variable to -1
	atomic.StoreInt32(&promise.waitCnt, -1)
	// set the commandline
	promise.cmd = *exec.CommandContext(promise.ctx, arg0, args...)
	promise.cmd.Stdin = inStream
	promise.cmd.Stdout = outStream
	promise.cmd.Stderr = outStream

	// set the error for run commandline goroutine
	promise.eg, _ = errgroup.WithContext(promise.ctx)
	promise.eg.Go(func() error {
		return promise.cmd.Run()
	})
	// set the conditional variable to 0
	atomic.StoreInt32(&promise.waitCnt, 0)
	return promise
}

func (p *Promise) Wait() (err error) {
	// if errgroup is not initialized, return error
	if p.eg == nil {
		return berror.BoxerError{
			Code:   berror.InvalidOperation,
			Origin: fmt.Errorf("promise is not initialized"),
			Msg:    "error while execute promise.Wait()",
		}
	}
	// if conditional variable is not set to 0, return error
	if atomic.LoadInt32(&p.waitCnt) != 0 {
		return berror.BoxerError{
			Code:   berror.InvalidOperation,
			Msg:    fmt.Sprintf("error while execute %s %v promise.Wait()", p.cmd.Path, p.cmd.Args),
			Origin: fmt.Errorf("promise is not initialized [%d]", atomic.LoadInt32(&p.waitCnt)),
		}
	}
	// set the conditional variable to 1
	atomic.StoreInt32(&p.waitCnt, 1)
	// wait for the errgroup to finish
	err = p.eg.Wait()
	if err != nil {
		return berror.BoxerError{
			Code:   berror.InvalidOperation,
			Msg:    fmt.Sprintf("error while execute %s %v promise.Wait()", p.cmd.Path, p.cmd.Args),
			Origin: err,
		}
	}
	return nil
}

func (p *Promise) Cancel() (err error) {
	if p.cancel == nil {
		return berror.BoxerError{
			Code:   berror.InvalidOperation,
			Msg:    "error while execute promise.Cancel()",
			Origin: fmt.Errorf("promise is not initialized"),
		}
	}
	if atomic.LoadInt32(&p.waitCnt) != 0 {
		return berror.BoxerError{
			Code:   berror.InvalidOperation,
			Msg:    fmt.Sprintf("error while execute %s %v promise.Cancel()", p.cmd.Path, p.cmd.Args),
			Origin: fmt.Errorf("promise is not initialized [%d]", atomic.LoadInt32(&p.waitCnt)),
		}
	}
	// cancel the context
	p.cancel()
	return nil
}
