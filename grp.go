package grp

import (
	"errors"
)

var (
	ErrPoolNotRuning = errors.New("The pool is not running")
	ErrJobNotFunc    = errors.New("Generic worker not given a func()")
	ErrWorkerClosed  = errors.New("worker was closed")
	ErrJobTimedOut   = errors.New("job request timed out")
)

func New(n int, ctor func() Worker) *Pool {
	return newPool(n, ctor)
}

func NewFunc(n int, f func(interface{}) interface{}) *Pool {
	return newPool(n, func() Worker {
		return &closureWorker{
			processor: f,
		}
	})
}

func NewCallback(n int) *Pool {
	return newPool(n, func() Worker {
		return &callbackWorker{}
	})
}
