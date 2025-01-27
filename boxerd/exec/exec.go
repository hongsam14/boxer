package exec

import (
	berror "boxerd/error"
	"context"
	"fmt"
	"os"
	"os/exec"

	"golang.org/x/sync/errgroup"
)

type Promise struct {
	ctx    context.Context
	cancel context.CancelFunc
	cmd    exec.Cmd
	eg     *errgroup.Group
}

func Run(inStream *os.File, outStream *os.File, arg0 string, args ...string) (promise *Promise) {
	promise = new(Promise)
	promise.ctx, promise.cancel = context.WithCancel(context.Background())
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
	return promise
}

func (p *Promise) Wait() (err error) {
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

func (p *Promise) Cancel() {
	p.cancel()
}
