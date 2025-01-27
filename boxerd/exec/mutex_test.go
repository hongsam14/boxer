package exec_test

import (
	"boxerd/exec"
	"sync"
	"testing"
	"time"

	"math/rand"
)

func TestCushioningSingle(t *testing.T) {
	wait_sec := rand.Intn(10)
	cushion := exec.InitCushion(uint(wait_sec))
	cushion.Lock()
	start := time.Now()
	cushion.Release()
	cushion.Lock()
	elapsed := time.Since(start)
	cushion.Release()
	t.Logf("CushionTime: %v Elapsed time: %v\n", wait_sec, elapsed)
	if elapsed < time.Duration(wait_sec)*time.Second {
		t.Errorf("Cushion is not waiting for the period")
	}
}

func TestCushioningMultiple(t *testing.T) {
	wait_sec := rand.Intn(10) + 1
	cushion := exec.InitCushion(uint(wait_sec))
	t.Logf("CushionTime: %v\n", wait_sec)

	rootStartTime := time.Now()

	wait := sync.WaitGroup{}
	cushion.Lock()
	start := time.Now()
	cushion.Release()
	for i := 0; i < wait_sec; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()

			startTime := time.Since(rootStartTime)
			cushion.Lock()
			elapsed := time.Since(start)
			cushion.Release()
			//next timer start
			start = time.Now()
			t.Logf("Start time: %v Elapsed time: %v\n", startTime, elapsed)
			if elapsed < time.Duration(wait_sec)*time.Second {
				t.Errorf("Cushion is not waiting for the period")
			}
		}()
	}
	wait.Wait()
}

func TestDuplicateRelease(t *testing.T) {
	cushion := exec.InitCushion(1)
	cushion.Lock()
	cushion.Release()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	cushion.Release()
}

func TestReleaseBeforeLock(t *testing.T) {
	cushion := exec.InitCushion(1)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	cushion.Release()
}

func TestReleaseWhileWaiting(t *testing.T) {
	cushion := exec.InitCushion(1)
	cushion.Lock()
	cushion.Release()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	cushion.Release()
}
