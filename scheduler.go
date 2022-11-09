package pool

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

// Usage:
//    pool.Start()
//    scheduler = NewScheduler(pool)
//
//    taskAt, err := scheduler.RunTaskAt(10s, job_function)
//    ...
//    taskAfter, err := scheduler.RunTaskAfter(taskAt, job_function_2)
//
//    taskBefore, err := scheduler.RunTaskBefore(taskAfter, job_function_3)
//
//    scheduler.Stop()
//    pool.Stop()

type JobFunction func() (interface{}, error)

// ::TODO:: implement all required methods, WorkerTask is temporally only
// ::TODO:: would it be better to extend WorkerTask?
type SchedulerTask struct {
	WorkerTask

	//	next *SchedulerTask
	//	prev *SchedulerTask

	task     Task
	function JobFunction
	ticker   *time.Ticker
}

func (sht *SchedulerTask) startTaskAt(when time.Duration) <-chan time.Time {
	sht.ticker = time.NewTicker(when)

	return sht.ticker.C
}

type Scheduler struct {
	com         chan int
	channels    []<-chan time.Time
	tasks       []*SchedulerTask
	pool        Pool
	ready       atomic.Value
	terminating *sync.WaitGroup
}

func poll[T any](chs []<-chan T, cmd chan int) (int, error) {
	fmt.Printf("POLL: Startign poll for %d records\n", len(chs))
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
	fmt.Printf("POLL: Selected: chosen %v, value %v, ok %v\n", chosen, value, ok)
	if !ok {
		return chosen, errors.New("channel closed")
	}

	return chosen, nil
}

func NewScheduler(p Pool) (*Scheduler, error) {
	ch := make(chan int)

	sch := &Scheduler{
		com:         ch,
		channels:    make([]<-chan time.Time, 0),
		tasks:       make([]*SchedulerTask, 0),
		pool:        p,
		terminating: &sync.WaitGroup{},
	}
	sch.terminating.Add(1)
	starting := &sync.WaitGroup{}

	starting.Add(1)

	go func() {
		starting.Done()
		for {
			fmt.Printf("POLLER: Wait poll, channels %d\n", len(sch.channels))
			selected, err := poll(sch.channels, sch.com)
			fmt.Printf("POLLER: Polled: selected %v, err %v\n", selected, err)
			fmt.Printf("POLLER: Len: %v\n", len(sch.channels))

			if selected == len(sch.channels) {
				fmt.Println("POLLER: It is comm")
				// ch channel, closed
				if err != nil {
					break // need to close all channels
				}
				// skip value on com channel;
				// probably need to rebuild poll
				continue
			}

			//time to run value
			if err == nil {
				fmt.Println("POLLER: set task to execute")
				scheduler_task := sch.tasks[selected]
				scheduler_task.ticker.Stop()

				// Delete: ::TODO:: should it be rewritten??
				sch.tasks = append(sch.tasks[:selected], sch.tasks[selected+1:]...)
				sch.channels = append(sch.channels[:selected], sch.channels[selected+1:]...)
				t, err := sch.pool.PushTask(scheduler_task.function)
				if err != nil {
					// ::FIME:: handle error
					fmt.Printf("ERROR: %v\n", err) // ::FIXME:: set error in the task
				}
				scheduler_task.task = t
				// ::FIXME:: there is new task returned, map it
			}
		}
		// close(ch) <- closed by stop

		sch.terminating.Done()
		fmt.Println("POLLER: Service func is quit")
	}()

	sch.ready.Store(true)

	starting.Wait() // Make sure everything is started

	fmt.Println("Scheduler: Init DONE")

	return sch, nil
}

func (sc *Scheduler) RunTaskAt(at time.Duration, job func() (interface{}, error)) (Task, error) {
	task := &SchedulerTask{}
	task.function = job

	task_channel := task.startTaskAt(at)

	fmt.Println("Scheduler: Task's tiket created")
	sc.channels = append(sc.channels, task_channel)
	sc.tasks = append(sc.tasks, task)

	fmt.Println("Scheduler: Notify poll")
	sc.com <- len(sc.tasks)
	fmt.Println("Scheduler: Notified")

	return task, nil
}

func (sc *Scheduler) Stop() bool {
	status := sc.ready.CompareAndSwap(true, false)
	if status {
		close(sc.com)
		sc.terminating.Wait()
	}
	return status
}

// func (sc *Scheduler) RunTaskBefore(before Task, job func() (interface{}, error)) (Task, error) {
// 	return nil, fmt.Errorf("Not Implemented")
// }

// func (sc *Scheduler) RunTaskAfter(after Task, job func() (interface{}, error)) (Task, error) {
// 	return nil, fmt.Errorf("Not Implemented")
// }
