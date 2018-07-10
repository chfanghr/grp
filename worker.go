package grp

type Worker interface {
	Process(interface{}) interface{}
	BlockUntilReady()
	Interrupt()
	Terminate()
}

type workRequest struct {
	jobChan       chan<- interface{}
	retChan       <-chan interface{}
	interruptFunc func()
}

type workerWrapper struct {
	worker        Worker
	interruptChan chan struct{}
	reqChan       chan<- workRequest
	closeChan     chan struct{}
	closedChan    chan struct{}
}

func (ww *workerWrapper) interrupt() {
	close(ww.interruptChan)
	ww.worker.Interrupt()
}

func (ww *workerWrapper) run() {
	jobChan, retChan := make(chan interface{}), make(chan interface{})
	defer func() {
		ww.worker.Terminate()
		close(retChan)
		close(ww.closedChan)
	}()
	for {
		ww.worker.BlockUntilReady()
		select {
		case ww.reqChan <- workRequest{
			jobChan:       jobChan,
			retChan:       retChan,
			interruptFunc: ww.interrupt,
		}:
			select {
			case payload := <-jobChan:
				result := ww.worker.Process(payload)
				select {
				case retChan <- result:
				case <-ww.interruptChan:
					ww.interruptChan = make(chan struct{})
				}
			case _, _ = <-ww.interruptChan:
				ww.interruptChan = make(chan struct{})
			}
		case <-ww.closeChan:
			return
		}
	}
}

func (ww *workerWrapper) stop() {
	close(ww.closeChan)
}

func (ww *workerWrapper) join() {
	<-ww.closedChan
}

func newWorkerWrapper(reqChan chan<- workRequest, worker Worker) *workerWrapper {
	ww := workerWrapper{
		worker:        worker,
		interruptChan: make(chan struct{}),
		reqChan:       reqChan,
		closeChan:     make(chan struct{}),
		closedChan:    make(chan struct{}),
	}
	go ww.run()
	return &ww
}
