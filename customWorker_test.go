package grp

import (
	"testing"
	"time"
)

func TestCustomWorker(t *testing.T) {
	pool := New(1, func() Worker {
		return &mockWorker{
			blockProcChan:  make(chan struct{}),
			blockReadyChan: make(chan struct{}),
			interruptChan:  make(chan struct{}),
		}
	})

	worker1, ok := pool.workers[0].worker.(*mockWorker)

	if !ok {
		t.Fatal("Wrong typpe of worker in pool")
	}

	if worker1.terminated {
		t.Fatal("Worker started off terminated")
	}

	_, err := pool.ProcessTimed(10, time.Millisecond)
	if exp, act := ErrJobTimedOut, err; exp != act {
		t.Errorf("Wrong error: %v != %v", act, exp)
	}

	close(worker1.blockReadyChan)
	_, err = pool.ProcessTimed(10, time.Millisecond)
	if exp, act := ErrJobTimedOut, err; exp != act {
		t.Errorf("Wrong error: %v != %v", act, exp)
	}

	close(worker1.blockProcChan)
	if exp, act := 10, pool.Process(10).(int); exp != act {
		t.Errorf("Wrong result: %v != %v", act, exp)
	}

	pool.Close()
	if !worker1.terminated {
		t.Fatal("Worker was not terminated")
	}
}
