package pool

// ::TODO:: use context to cancel working tasks ?

import (
	"fmt"
	"sync"
)

type WorkStatus int

const (
	Available WorkStatus = iota
	Waiting
	Working
	Finished
	Error
)

type WorkerTask struct {
	lock   sync.Mutex
	status WorkStatus
	id     int
	result interface{}
	error  error

	comm chan func() (interface{}, error)
	cond *sync.Cond
}

func createWorkerTask(id int) (*WorkerTask, error) {
	task := WorkerTask{
		lock:   sync.Mutex{},
		status: Available,
		id:     id,
		result: nil,
		error:  nil,
		comm:   make(chan func() (interface{}, error)),
		cond:   nil,
	}

	task.cond = sync.NewCond(&task.lock)
	return &task, nil
}

func (wt *WorkerTask) getStatus() WorkStatus {
	wt.lock.Lock()
	defer wt.lock.Unlock()
	status := wt.status

	return status
}

func (wt *WorkerTask) setStatus(status WorkStatus, result interface{}, err error) {
	wt.lock.Lock()
	defer wt.lock.Unlock()

	wt.status = status
	wt.result = result
	wt.error = err

	wt.cond.Broadcast()
}

func (wt *WorkerTask) IsWaiting() bool {
	return wt.getStatus() == Waiting
}

func (wt *WorkerTask) IsFinished() bool {
	return wt.getStatus() == Waiting
}

func (wt *WorkerTask) IsWorking() bool {
	return wt.getStatus() == Working
}

func (wt *WorkerTask) IsError() bool {
	return wt.getStatus() == Error
}

func (wt *WorkerTask) Result() (interface{}, error) {
	wt.lock.Lock()
	defer wt.lock.Unlock()
	return wt.result, wt.error
}

func (wt *WorkerTask) WaitFinished() (interface{}, error) {
	wt.cond.L.Lock()
	if wt.status != Working {
		wt.cond.L.Unlock()
		return nil, fmt.Errorf("Worng status")
	}

	for {
		if wt.status != Finished {
			wt.cond.Wait()
		} else {
			break
		}
	}
	defer wt.cond.L.Unlock()

	return wt.result, wt.error
}
