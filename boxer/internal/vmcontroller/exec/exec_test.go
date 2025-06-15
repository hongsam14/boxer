package exec_test

import (
	"boxerd/internal/vmcontroller/exec"
	"os"
	"testing"
	"time"
)

func TestPromise(t *testing.T) {
	promise, err := exec.Run(os.Stdin, os.Stdout, "echo", "hello")
	if err != nil {
		t.Errorf("Error while executing promise %v", err)
		return
	}
	_, err = promise.Wait()
	if err != nil {
		t.Errorf("Error while executing promise %v", err)
	}
	if !promise.IsExecuted() {
		t.Errorf("Promise should be executed")
		return
	}
}

func TestPromiseError(t *testing.T) {
	_, err := exec.Run(os.Stdin, os.Stdout, "asdf", "ghjk")
	if err == nil {
		t.Errorf("Error should be raised executing promise %v", err)
		return
	}
}

// (DEPRECATED) Input is not supported anymore
// func TestPromiseInput(t *testing.T) {
// 	promise, err := exec.Run(os.Stdin, os.Stdout, "man", "watch")
// 	if err != nil {
// 		t.Errorf("Error while executing promise %v", err)
// 		return
// 	}
// 	_, err = promise.Wait()
// 	if err != nil {
// 		t.Errorf("Error while executing promise %v", err)
// 	}
// }

func TestPromiseWithoutWait(t *testing.T) {
	start := time.Now()

	promise, err := exec.Run(os.Stdin, os.Stdout, "watch", "-n", "1", "echo", "hello")
	if err != nil {
		t.Errorf("Error while executing promise %v", err)
		return
	}
	defer promise.Cancel()

	// check is the promise is running
	if !promise.IsExecuted() {
		t.Errorf("Promise should be executed")
		return
	}

	elapsed := time.Since(start)
	if elapsed > 1*time.Second {
		t.Errorf("Promise should not have waited for the command to finish")
	}
}

// (DEPRECATED) Promise is interface so this test is not needed anymore
// func TestWaitBeforePromise(t *testing.T) {
// 	promise := new(exec.Promise)
// 	_, err := promise.Wait()
// 	if err == nil {
// 		t.Errorf("Error should have been raised")
// 	}
// 	t.Logf("Error: %v", err)
// }

// (DEPRECATED) Promise is interface so this test is not needed anymore
// func TestCancelBeforePromise(t *testing.T) {
// 	promise := new(exec.Promise)
// 	_, err := promise.Cancel()
// 	if err == nil {
// 		t.Errorf("Error should have been raised")
// 	}
// 	t.Logf("Error: %v", err)
// }

func TestWaitFuncCallDuplicated(t *testing.T) {
	promise, err := exec.Run(os.Stdin, os.Stdout, "echo", "hello")
	if err != nil {
		t.Errorf("Error while executing promise %v", err)
		return
	}
	_, err = promise.Wait()
	if err != nil {
		t.Errorf("Error while executing promise %v", err)
	}
	_, err = promise.Wait()
	if err == nil {
		t.Errorf("Error should have been raised")
	}
	t.Logf("Error: %v", err)
}
