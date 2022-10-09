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

	comm chan func() error
}

func createWorkerTask(id int) (*WorkerTask, error) {
	task := WorkerTask{
		status: Available,
		id:     id,
		comm:   make(chan func() error),
	}
	return &task, nil
}

func (wt *WorkerTask) setStatus(status WorkStatus) {
	wt.lock.Lock()
	defer wt.lock.Unlock()
	wt.status = status

}

type WorkerPool struct {
	max_size  int
	tasks     map[int]*WorkerTask
	available []*WorkerTask
}

func NewWorkerPool(size int) (*WorkerPool, error) {

	return &WorkerPool{
		max_size:  size,
		available: nil,
	}, nil
}

func (w *WorkerPool) Start() error {
	w.available = make([]*WorkerTask, w.max_size)

	for i := 0; i < w.max_size; i++ {
		task, err := createWorkerTask(i)
		if err != nil {
			return err
		}
		w.tasks[i] = task
		w.available[i] = task
		go func(task *WorkerTask) {
			for {
				task.setStatus(Available)

				worker := <-task.comm
				if worker == nil {
					break
				}
				task.setStatus(Working)
				// ::TODO:: move to busy list
				err := worker()
				// ::TODO:: Move to available list
				if err != nil {
					task.setStatus(Error)
				} else {
					task.setStatus(Finished)

				}
			}
		}(task)

	}
	return nil
}

func (w *WorkerPool) Stop() error {
	// ::TODO::
	// - check the status of the task
	// - if task is working return error
	for i := range w.tasks {
		w.tasks[i].comm <- nil
	}

	return nil
}

func (w *WorkerPool) PushTask(f func() error) error {
	// ::TODO:: allow grow if max allow it

	// Should it be lock here?
	if len(w.available) == 0 {
		return fmt.Errorf("Pool is overloaded")
	}

	if len(w.available) > 0 {
		var t *WorkerTask
		t, w.available = w.available[0], w.available[1:]
		//t, w.available := w.available[len(w.available)-1], w.available[:len(w.available)-1]
		t.comm <- f
	}

	return nil
}
