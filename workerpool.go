package pool

// ::TODO:: use context to cancel working tasks ?

import (
	"fmt"
	"sync"
)

type WorkerPool struct {
	lock sync.Mutex

	max_size  int
	tasks     map[int]*WorkerTask
	available []*WorkerTask
}

func NewWorkerPool(size int) (*WorkerPool, error) {

	return &WorkerPool{
		lock:     sync.Mutex{},
		max_size: size,

		available: nil,
	}, nil
}

func (w *WorkerPool) markTaskBusy(task *WorkerTask) {
	w.lock.Lock()
	defer w.lock.Unlock()

	for i, e := range w.available {
		if e.id == task.id {
			w.available = append(w.available[:i], w.available[i+1:]...)
			break
		}
	}

	task.setStatus(Working, nil, nil)
}

func (w *WorkerPool) markTaskAvailable(task *WorkerTask) error {
	if task.getStatus() == Available {
		return nil
	}

	w.lock.Lock()
	defer w.lock.Unlock()

	_, ok := w.tasks[task.id]
	if !ok {
		return fmt.Errorf("Bad task")
	}

	task.setStatus(Available, nil, nil)

	w.available = append(w.available, task)
	return nil
}

func (w *WorkerPool) Start() error {
	w.available = make([]*WorkerTask, w.max_size)
	w.tasks = make(map[int]*WorkerTask)

	for i := 0; i < w.max_size; i++ {
		task, err := createWorkerTask(i)
		if err != nil {
			return err
		}
		w.tasks[i] = task
		w.available[i] = task
		go func(task *WorkerTask) {
			for {
				task.setStatus(Available, nil, nil)

				worker := <-task.comm
				if worker == nil {
					break
				}
				w.markTaskBusy(task)
				res, err := worker()
				if err != nil {
					task.setStatus(Error, res, err)
				} else {
					task.setStatus(Finished, res, nil)
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

func (w *WorkerPool) PushTask(f func() (interface{}, error)) (Task, error) {
	// ::TODO:: allow grow if max allows it

	// Should it be lock-ed here?
	if len(w.available) == 0 {
		return nil, fmt.Errorf("Pool is overloaded")
	}

	var t *WorkerTask

	if len(w.available) > 0 {

		t, w.available = w.available[0], w.available[1:]
		//t, w.available := w.available[len(w.available)-1], w.available[:len(w.available)-1]
		t.comm <- f
	}

	return t, nil
}

func (w *WorkerPool) ReleaseTask(task Task) error {
	working_task, ok := task.(*WorkerTask)
	if !ok {
		return fmt.Errorf("Wrong task type")
	}

	return w.markTaskAvailable(working_task)
}
