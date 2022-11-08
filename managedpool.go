package pool

// ::TODO:: test case for long run& pool stop immediately
// ::TODO:: use context to cancel working tasks ?

import (
	"fmt"
	"sync"
)

type ManagedPool struct {
	lock        sync.Mutex
	max_size    int
	tasks       map[int]*WorkerTask
	available   []*WorkerTask
	terminating *sync.WaitGroup
	ucond       *sync.Cond
}

func NewManagedPool(size int) (*ManagedPool, error) {

	return &ManagedPool{
		lock:        sync.Mutex{},
		max_size:    size,
		tasks:       nil,
		available:   nil,
		terminating: nil,
	}, nil
}

func (w *ManagedPool) markTaskBusy(task *WorkerTask) {
	if task.IsAvailable() {
		w.lock.Lock()
		defer w.lock.Unlock()

		for i, e := range w.available {
			if e.id == task.id {
				w.available = append(w.available[:i], w.available[i+1:]...)
				break
			}
		}
	}

	task.setStatus(Working, nil, nil)
}

func (w *ManagedPool) markTaskTerminated(task *WorkerTask) {
	w.lock.Lock()
	defer w.lock.Unlock()

	for i, e := range w.available {
		if e.id == task.id {
			w.available = append(w.available[:i], w.available[i+1:]...)
			break
		}
	}

	task.setStatus(Finished, nil, nil)

	delete(w.tasks, task.id)
	w.terminating.Done()
}

func (w *ManagedPool) markTaskAvailable(task *WorkerTask) error {
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

func (w *ManagedPool) Start() error {
	w.available = make([]*WorkerTask, w.max_size)
	w.tasks = make(map[int]*WorkerTask)
	//w.starting = &sync.WaitGroup{}
	w.terminating = &sync.WaitGroup{}
	w.ucond = sync.NewCond(&w.lock)

	//w.starting.Add(w.max_size)

	for i := 0; i < w.max_size; i++ {
		task, err := createWorkerTask(i)
		if err != nil {
			return err
		}
		w.tasks[i] = task
		w.available[i] = task

		w.terminating.Add(1)
		task.setStatus(Available, nil, nil)

		go func(task *WorkerTask) {
			//w.starting.Done()
			for {
				job := <-task.comm
				if job == nil {
					w.markTaskTerminated(task)
					break
				}
				w.markTaskBusy(task)
				res, err := job()
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

func (w *ManagedPool) Stop() error {
	// ::TODO::
	// - should be error returned in case when task(s) busy?
	if w.terminating == nil {
		return nil // not started
	}
	{
		w.lock.Lock()

		for _, el := range w.tasks {
			ch := w.tasks[el.id].comm
			w.lock.Unlock()

			ch <- nil

			w.lock.Lock()
		}
		w.lock.Unlock()
	}
	w.terminating.Wait()

	return nil
}

func (w *ManagedPool) PushTask(f func() (interface{}, error)) (Task, error) {
	// ::TODO:: allow grow if max allows it

	w.lock.Lock()
	defer w.lock.Unlock()

	if len(w.available) == 0 {
		return nil, fmt.Errorf("Pool is overloaded")
	}

	var t *WorkerTask

	t, w.available = w.available[0], w.available[1:]
	//t, w.available := w.available[len(w.available)-1], w.available[:len(w.available)-1]
	t.setStatus(Working, nil, nil)
	//t.setManageredStatus(true)
	t.comm <- f

	return t, nil
}

func (w *ManagedPool) ReleaseTask(task Task) error {
	working_task, ok := task.(*WorkerTask)
	if !ok {
		return fmt.Errorf("Wrong task type")
	}

	if working_task.IsAvailable() {
		return fmt.Errorf("Wrong task state")
	}

	return w.markTaskAvailable(working_task)
}
