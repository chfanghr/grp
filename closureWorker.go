package grp

type closureWorker struct {
	processor func(interface{}) interface{}
}

func (cw *closureWorker) Process(payload interface{}) interface{} {
	return cw.processor(payload)
}

func (cw *closureWorker) BlockUntilReady() {}
func (cw *closureWorker) Interrupt()       {}
func (cw *closureWorker) Terminate()       {}
