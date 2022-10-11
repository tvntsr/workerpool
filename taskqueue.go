package pool

import (
	"fmt"
	"sync"
	"time"
)

type WorkerTaskQueue struct {
	lock     sync.Mutex
	elements map[int]*WorkerTask
	comm     chan int
	quit     chan interface{}
}

func NewWorkerTaskQueue() *WorkerTaskQueue {
	return &WorkerTaskQueue{
		lock:     sync.Mutex{},
		elements: make(map[int]*WorkerTask),
		comm:     make(chan int),
		quit:     make(chan interface{}),
	}
}

func (wtq *WorkerTaskQueue) getTaskWithState(id int) (*WorkerTask, error) {
	wtq.lock.Lock()
	defer wtq.lock.Unlock()

	task, ok := wtq.elements[id]
	if !ok {
		return nil, fmt.Errorf("Element %d not found", id)

	}

	task_status := task.getStatus()
	if task_status == Finished || task_status == Error {
		task.setStatusChan(nil)
		// Lock already here
		delete(wtq.elements, id)

		return task, nil
	}

	return nil, nil
}

func (wtq *WorkerTaskQueue) TaimedWaitTaskFinished(timeout time.Duration) (*WorkerTask, error) {
	for {
		select {
		case id := <-wtq.comm:
			task, err := wtq.getTaskWithState(id)
			if task != nil || err != nil {
				return task, err
			}
		case <-time.After(timeout):
			return nil, fmt.Errorf("timeout")
		case <-wtq.quit:
			return nil, nil
		}
	}

	// Unreachable
	//return nil, nil
}

func (wtq *WorkerTaskQueue) WaitTaskFinished() (*WorkerTask, error) {
Loop:
	for {
		select {
		case id := <-wtq.comm:
			task, err := wtq.getTaskWithState(id)
			if task != nil || err != nil {
				return task, err
			}
		case <-wtq.quit:
			break Loop
		}
	}

	return nil, nil
}

func (wtq *WorkerTaskQueue) AddTask(wt *WorkerTask) bool {
	wtq.lock.Lock()
	defer wtq.lock.Unlock()

	var ok bool

	if _, ok = wtq.elements[wt.id]; !ok {
		wtq.elements[wt.id] = wt
		wt.setStatusChan(&wtq.comm)
	}

	return !ok
}

func (wtq *WorkerTaskQueue) RemoveTask(wt *WorkerTask) bool {
	wtq.lock.Lock()
	defer wtq.lock.Unlock()

	var ok bool

	if _, ok = wtq.elements[wt.id]; ok {
		delete(wtq.elements, wt.id)
		wt.setStatusChan(nil)
	}

	return ok
}
