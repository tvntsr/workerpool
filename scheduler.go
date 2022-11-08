package pool

import (
	"errors"
	"fmt"
	"reflect"
	"sync/atomic"
	"time"
)

/*
Usage:
   pool.Start()
   scheduler = NewScheduler(pool)

   taskAt, err := scheduler.RunTaskAt(10s, job_function)
   ...
   taskAfter, err := scheduler.RunTaskAfter(taskAt, job_function_2)

   taskBefore, err := scheduler.RunTaskBefore(taskAfter, job_function_3)

   scheduler.Stop()
   pool.Stop()
*/

type JobFunction func() (interface{}, error)

type SchedulerTask struct {
	next *SchedulerTask
	prev *SchedulerTask

	task Task
}

func (sht *SchedulerTask) setTask(t Task) {
	sht.task = t
}

type Scheduler struct {
	ch    chan JobFunction
	chs   []chan JobFunction
	pool  Pool
	ready atomic.Value
}

func poll[T any](chs []chan T, cmd chan JobFunction) (int, T, error) {
	var zeroT T

	cases := make([]reflect.SelectCase, len(chs)+1)

	for i, ch := range chs {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}

	cases[len(chs)] = reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(cmd),
	}

	// ok will be true if the channel has not been closed.
	chosen, value, ok := reflect.Select(cases)
	if !ok {
		return chosen, zeroT, errors.New("channel closed")
	}
	if ret, ok := value.Interface().(T); ok {
		return chosen, ret, nil
	}

	return chosen, zeroT, errors.New("failed to cast value")
}

func NewScheduler(p Pool) (*Scheduler, error) {
	ch := make(chan JobFunction)

	sch := &Scheduler{
		ch:   ch,
		chs:  make([]chan JobFunction, 0),
		pool: p,
	}

	go func() {
		for {
			selected, value, err := poll(sch.chs, sch.ch)
			if selected == len(sch.chs)+1 {
				// ch channel, closed
				if err != nil {
					break // need to close all channels
				}
				// skip value on com channel;
				// probably need to rebuild poll
				continue
			}

			// time to run value
			if err == nil {
				sch.pool.PushTask(value)
			}
		}
		// close all channels
		//close(ch) <- closed by stop
		for _, c := range sch.chs {
			close(c)
		}
	}()

	sch.ready.Store(true)

	return sch, nil
}

func (sc *Scheduler) Stop() bool {
	status := sc.ready.CompareAndSwap(true, false)
	if status {
		close(sc.ch)
	}
	return status
}

func (sc *Scheduler) RunTaskAt(at time.Duration, job func() (interface{}, error)) (Task, error) {
	task := &SchedulerTask{}

	f := func() (interface{}, error) {
		ticker := time.NewTicker(at)
		done := make(chan bool)
		select {
		case <-done:
			ticker.Stop()
			return nil, fmt.Errorf("Cancelled")
		case <-ticker.C:
			ticker.Stop()
			return job()
		}
	}

	t, err := sc.pool.PushTask(f)
	if err != nil {
		return nil, err
	}

	task.setTask(t)

	return nil, fmt.Errorf("Not Implemented")
}

// func (sc *Scheduler) RunTaskBefore(before Task, job func() (interface{}, error)) (Task, error) {
// 	return nil, fmt.Errorf("Not Implemented")
// }

// func (sc *Scheduler) RunTaskAfter(after Task, job func() (interface{}, error)) (Task, error) {
// 	return nil, fmt.Errorf("Not Implemented")
// }
