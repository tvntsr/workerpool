package pool

// ::TODO:: use context to cancel working tasks ?

import (
	"sync"
)

type UnmanagedPool struct {
	jobs        chan func() (interface{}, error)
	lock        sync.Mutex
	max_size    int
	tasks       map[int]*WorkerTask
	terminating *sync.WaitGroup
	unmanaged   []func() (interface{}, error)
}

func NewUnmanagedPool(size int) (*UnmanagedPool, error) {

	return &UnmanagedPool{
		lock:        sync.Mutex{},
		max_size:    size,
		tasks:       nil,
		terminating: nil,
		unmanaged:   nil,
	}, nil
}

func (w *UnmanagedPool) markTaskBusy(task *WorkerTask) {

	task.setStatus(Working, nil, nil)
}

func (w *UnmanagedPool) markTaskTerminated(task *WorkerTask) {
	w.lock.Lock()
	defer w.lock.Unlock()

	task.setStatus(Finished, nil, nil)

	delete(w.tasks, task.id)
	w.terminating.Done()
}

func (w *UnmanagedPool) getUnmanageredWork() func() (interface{}, error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	if w.unmanaged == nil || len(w.unmanaged) == 0 {
		return nil
	}

	var t func() (interface{}, error)
	if len(w.unmanaged) == 1 {
		t = w.unmanaged[0]
		w.unmanaged = make([]func() (interface{}, error), 0)
	} else {
		t, w.unmanaged = w.unmanaged[0], w.unmanaged[1:]
	}

	return t
}

func (w *UnmanagedPool) Start() error {
	w.jobs = make(chan func() (interface{}, error))
	w.tasks = make(map[int]*WorkerTask)
	w.terminating = &sync.WaitGroup{}

	w.unmanaged = make([]func() (interface{}, error), 0)

	w.terminating.Add(w.max_size)
	for i := 0; i < w.max_size; i++ {
		task, err := createWorkerTask(i)
		if err != nil {
			return err
		}
		w.tasks[i] = task

		task.setStatus(Available, nil, nil)

		go func(task *WorkerTask) {
			for {
				worker, ok := <-w.jobs
				if !ok || worker == nil {
					w.markTaskTerminated(task)
					break
				}
				w.markTaskBusy(task)
				_, _ = worker()

				t := w.getUnmanageredWork()
				if t == nil {
					continue
				}

				select {
				case w.jobs <- t:
					continue
				default:
					_, _ = t()
				}
			}
		}(task)

	}
	return nil
}

func (w *UnmanagedPool) Stop() error {
	// ::TODO::
	// - should be error returned in case when task(s) busy?
	if w.terminating == nil {
		return nil // not started
	}

	{
		close(w.jobs)
	}

	w.terminating.Wait()

	return nil
}

func (w *UnmanagedPool) PushTask(f func() (interface{}, error)) (Task, error) {

	select {
	case w.jobs <- f:
		return nil, nil
	default:
		w.lock.Lock()
		defer w.lock.Unlock()

		w.unmanaged = append(w.unmanaged, f)
	}

	return nil, nil
}

func (w *UnmanagedPool) ReleaseTask(task Task) error {
	return nil
}
