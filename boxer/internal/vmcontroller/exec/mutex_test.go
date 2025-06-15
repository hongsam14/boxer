package exec_test

import (
	"boxerd/internal/vmcontroller/exec"
	"sync"
	"testing"
	"time"

	"math/rand"
)

func TestPaddedMutexingSingle(t *testing.T) {
	wait_sec := rand.Intn(10)
	pMux := exec.InitPaddedMutex(uint(wait_sec))
	pMux.Lock()
	start := time.Now()
	pMux.Release()
	pMux.Lock()
	elapsed := time.Since(start)
	pMux.Release()
	t.Logf("PaddedMutexTime: %v Elapsed time: %v\n", wait_sec, elapsed)
	if elapsed < time.Duration(wait_sec)*time.Second {
		t.Errorf("PaddedMutex is not waiting for the period")
	}
}

func TestPaddedMutexingMultiple(t *testing.T) {
	wait_sec := rand.Intn(10) + 1
	pMux := exec.InitPaddedMutex(uint(wait_sec))
	t.Logf("PaddedMutexTime: %v\n", wait_sec)

	rootStartTime := time.Now()

	wait := sync.WaitGroup{}
	pMux.Lock()
	start := time.Now()
	pMux.Release()
	for i := 0; i < wait_sec; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()

			startTime := time.Since(rootStartTime)
			pMux.Lock()
			elapsed := time.Since(start)
			pMux.Release()
			//next timer start
			start = time.Now()
			t.Logf("Start time: %v Elapsed time: %v\n", startTime, elapsed)
			if elapsed < time.Duration(wait_sec)*time.Second {
				t.Errorf("PaddedMutex is not waiting for the period")
			}
		}()
	}
	wait.Wait()
}

func TestDuplicateRelease(t *testing.T) {
	pMux := exec.InitPaddedMutex(1)
	pMux.Lock()
	pMux.Release()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	pMux.Release()
}

func TestReleaseBeforeLock(t *testing.T) {
	pMux := exec.InitPaddedMutex(1)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	pMux.Release()
}

func TestReleaseWhileWaiting(t *testing.T) {
	pMux := exec.InitPaddedMutex(1)
	pMux.Lock()
	pMux.Release()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	pMux.Release()
}
