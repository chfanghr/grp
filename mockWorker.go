package grp

type mockWorker struct {
	blockProcChan  chan struct{}
	blockReadyChan chan struct{}
	interruptChan  chan struct{}
	terminated     bool
}

func (m *mockWorker) Process(in interface{}) interface{} {
	select {
	case <-m.blockProcChan:
	case <-m.interruptChan:
	}
	return in
}

func (m *mockWorker) BlockUntilReady() {
	<-m.blockReadyChan
}

func (m *mockWorker) Interrupt() {
	m.terminated = true
}

func (m *mockWorker) Terminate() {
	m.terminated = true
}
