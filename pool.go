package grp

import (
	"sync"
	"sync/atomic"
	"time"
)

type Pool struct {
	ctor    func() Worker
	workers []*workerWrapper
	reqChan chan workRequest

	workerMut sync.Mutex
	queueJobs int64
}

func (p *Pool) Process(payload interface{}) interface{} {
	atomic.AddInt64(&p.queueJobs, 1)

	request, open := <-p.reqChan
	if !open {
		panic(ErrPoolNotRuning)
	}

	request.jobChan <- payload

	payload, open = <-request.retChan
	if !open {
		panic(ErrWorkerClosed)
	}

	atomic.AddInt64(&p.queueJobs, -1)
	return payload
}

func (p *Pool) ProcessTimed(payload interface{}, timeout time.Duration) (interface{}, error) {
	atomic.AddInt64(&p.queueJobs, 1)
	defer atomic.AddInt64(&p.queueJobs, -1)

	tout := time.NewTimer(timeout)
	var request workRequest
	var open bool

	select {
	case request, open = <-p.reqChan:
		if !open {
			return nil, ErrJobTimedOut
		}
	case <-tout.C:
		return nil, ErrJobTimedOut
	}

	select {
	case request.jobChan <- payload:
	case <-tout.C:
		request.interruptFunc()
		return nil, ErrJobTimedOut
	}

	select {
	case payload, open = <-request.retChan:
		if !open {
			return nil, ErrWorkerClosed
		}
	case <-tout.C:
		request.interruptFunc()
		return nil, ErrJobTimedOut
	}

	tout.Stop()
	return payload, nil
}

func (p *Pool) QueueLength() int64 {
	return atomic.LoadInt64(&p.queueJobs)
}

func (p *Pool) SetSize(n int) {
	p.workerMut.Lock()
	defer p.workerMut.Unlock()
	lWorkers := len(p.workers)

	if lWorkers == n {
		return
	}

	for i := lWorkers; i < n; i++ {
		p.workers = append(p.workers, newWorkerWrapper(p.reqChan, p.ctor()))
	}

	for i := n; i < lWorkers; i++ {
		p.workers[i].stop()
	}

	for i := n; i < lWorkers; i++ {
		p.workers[i].join()
	}

	p.workers = p.workers[:n]
}

func (p *Pool) GetSize() int {
	p.workerMut.Lock()
	defer p.workerMut.Unlock()
	return len(p.workers)
}

func (p *Pool) Close() {
	p.SetSize(0)
	close(p.reqChan)
}

func newPool(n int, ctor func() Worker) *Pool {
	p := &Pool{
		ctor:    ctor,
		reqChan: make(chan workRequest),
	}
	p.SetSize(n)
	return p
}
