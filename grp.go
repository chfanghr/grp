package grp

import (
	"errors"
)

/*一些由grp定义和使用的error*/
var (
	ErrPoolNotRuning = errors.New("The pool is not running")
	ErrJobNotFunc    = errors.New("Generic worker not given a func()")
	ErrWorkerClosed  = errors.New("worker was closed")
	ErrJobTimedOut   = errors.New("job request timed out")
)

/*
New 返回新建的goroutine池指针
n goroutine数
ctor 提供worker的构造函数
*/
func New(n int, ctor func() Worker) *Pool {
	return newPool(n, ctor)
}

/*
NewFunc 返回新建的goroutine池指针,由函数作为worker goroutine
n goroutine数
f 提供函数作为worker
池中的goroutine不能被中断或停止,BlockUntilReady方法不阻塞
*/
func NewFunc(n int, f func(interface{}) interface{}) *Pool {
	return newPool(n, func() Worker {
		return &closureWorker{
			processor: f,
		}
	})
}

/*
NewCallback 返回新建的回调池指针
n goroutine数
池中的goroutine不能被中断或停止,BlockUntilReady方法不阻塞
*/
func NewCallback(n int) *Pool {
	return newPool(n, func() Worker {
		return &callbackWorker{}
	})
}
