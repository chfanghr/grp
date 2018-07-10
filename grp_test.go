package grp

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPoolSizeAdjustment(t *testing.T) {
	pool := NewFunc(10, func(interface{}) interface{} {
		return "foo"
	})
	if exp, act := 10, len(pool.workers); exp != act {
		t.Errorf("Wrong size of pool: %v != %v", act, exp)
	}
	pool.SetSize(10)
	if exp, act := 10, pool.GetSize(); exp != act {
		t.Errorf("Wrong size of pool: %v != %v", act, exp)
	}
	pool.SetSize(9)
	if exp, act := 9, pool.GetSize(); exp != act {
		t.Errorf("Wrong size of pool: %v != %v", act, exp)
	}
	pool.SetSize(10)
	if exp, act := 10, pool.GetSize(); exp != act {
		t.Errorf("Wrong size of pool: %v != %v", act, exp)
	}
	pool.SetSize(0)
	if exp, act := 0, pool.GetSize(); exp != act {
		t.Errorf("Wrong size of pool: %v != %v", act, exp)
	}
	pool.SetSize(10)
	if exp, act := 10, pool.GetSize(); exp != act {
		t.Errorf("Wrong size of pool: %v != %v", act, exp)
	}

	if exp, act := "foo", pool.Process(0).(string); exp != act {
		t.Errorf("Wrong result: %v != %v", act, exp)
	}

	pool.Close()
	if exp, act := 0, pool.GetSize(); exp != act {
		t.Errorf("Wrong size of pool: %v != %v", act, exp)
	}
}

func TestFuncJob(t *testing.T) {
	pool := NewFunc(10, func(in interface{}) interface{} {
		intVal := in.(int)
		return 2 * intVal
	})

	defer pool.Close()

	for i := 0; i < 10; i++ {
		ret := pool.Process(10)
		if exp, act := 20, ret.(int); exp != act {
			t.Errorf("Wrong result:%v!=%v", act, exp)
		}
	}
}

func TestFuncJobTimed(t *testing.T) {
	pool := NewFunc(10, func(in interface{}) interface{} {
		intVal := in.(int)
		return intVal * 2
	})
	defer pool.Close()

	for i := 0; i < 10; i++ {
		ret, err := pool.ProcessTimed(10, time.Millisecond)
		if err != nil {
			t.Fatalf("Failed to process: %v", err)
		}
		if exp, act := 20, ret.(int); exp != act {
			t.Errorf("Wrong result: %v != %v", act, exp)
		}
	}
}

func TestCallbackJob(t *testing.T) {
	pool := NewCallback(10)
	defer pool.Close()

	var counter int32
	for i := 0; i < 10; i++ {
		ret := pool.Process(func() {
			atomic.AddInt32(&counter, 1)
		})
		if ret != nil {
			t.Errorf("Non-nil callback response: %v", ret)
		}
	}

	ret := pool.Process("foo")
	if exp, act := ErrJobNotFunc, ret; exp != act {
		t.Errorf("Wrong result from non-func: %v != %v", act, exp)
	}

	if exp, act := int32(10), counter; exp != act {
		t.Errorf("Wrong result: %v != %v", act, exp)
	}
}

func TestParallelJobs(t *testing.T) {
	nWorkers := 10

	jobGroup := sync.WaitGroup{}
	testGroup := sync.WaitGroup{}

	pool := NewFunc(nWorkers, func(in interface{}) interface{} {
		jobGroup.Done()
		jobGroup.Wait()

		intVal := in.(int)
		return intVal * 2
	})
	defer pool.Close()

	for j := 0; j < 1; j++ {
		jobGroup.Add(nWorkers)
		testGroup.Add(nWorkers)

		for i := 0; i < nWorkers; i++ {
			go func() {
				ret := pool.Process(10)
				if exp, act := 20, ret.(int); exp != act {
					t.Errorf("Wrong result: %v != %v", act, exp)
				}
				testGroup.Done()
			}()
		}

		testGroup.Wait()
	}
}

func BenchmarkFuncJob(b *testing.B) {
	pool := NewFunc(10, func(in interface{}) interface{} {
		intVal := in.(int)
		return intVal * 2
	})
	defer pool.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ret := pool.Process(10)
		if exp, act := 20, ret.(int); exp != act {
			b.Errorf("Wrong result: %v != %v", act, exp)
		}
	}
}

func BenchmarkFuncTimedJob(b *testing.B) {
	pool := NewFunc(10, func(in interface{}) interface{} {
		intVal := in.(int)
		return intVal * 2
	})
	defer pool.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ret, err := pool.ProcessTimed(10, time.Second)
		if err != nil {
			b.Error(err)
		}
		if exp, act := 20, ret.(int); exp != act {
			b.Errorf("Wrong result: %v != %v", act, exp)
		}
	}
}
