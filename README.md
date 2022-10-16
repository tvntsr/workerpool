## Worker pool library
Sometimes it is needed to start different jobs in the pool and wait for the result

Despite it is relatively simple task in the Go, the proper implementation could take time because of synchronization

This library have two types of the task pool:
- Pool for managed task, when result of the job execution is required and task will be reused only when job results are consumed
- Pool for un-managed tasks, when result of the job execution is not important and task getting free to run other job without delay for result checks.

## Pool interface
Managed and un-managed pools are implementing pool interface:
```
type Pool interface {
        Start() error
        Stop() error
        PushTask(func() (interface{}, error)) (Task, error)
        ReleaseTask(Task) error
}
```

Idea behind the pool is the following usage:
```
  pool.Start()
  task, err := working_pool.PushTask(job)
  
  for {
      if task.IsFinished() {
         working_pool.ReleaseTask(task)
         break
      }
  }
  
  pool.Stop()
```

With WorkerTaskQueue it is possible to wait any task done:
```
  // checking err is skipped
  queue := NewWorkerTaskQueue()

  pool.Start()
  task, err := working_pool.PushTask(job)
  queue.AddTAsk(task)
  // Add n tasks to the pool ...
  task_n, err := working_pool.PushTask(job_n)
  queue.AddTAsk(task_n)
  
  for (i:= 0; i < n; i++) {
        task_done, err  := queue.WaitTaskFinished()
        
        // check result of the task execution is skipped
        
        pool.ReleaseTask(task_done)
  }
  
  pool.Stop()

```

In case of un-managed tasks the followig approach could be used:
```
  pool.Start()
  
  _, err := working_pool.PushTask(job)
  // check result of the task execution is not needed as well as task releasing
  
  pool.Stop()

```

## Waiting task
WorkerTaskQueue allows wait task done. This is very useful in case of managed task and wating for the one or several tasks results.

### WorkerTaskQueue definition
The following methods are implemented in WorkerTaskQueue
  - NewWorkerTaskQueue()
  - TaimedWaitTaskFinished(timeout time.Duration) (*WorkerTask, error)
  - WaitTaskFinished() (*WorkerTask, error)
  - AddTask(wt *WorkerTask) bool
  - RemoveTask(wt *WorkerTask) bool

 ## WorkerPool
 Pool for the managed jobs, ie result of the execution is important and task will be reused only if the results are retrieved
 
 ## UnmanagedPool
 Pool for the un-managed jobs, ie result of the execution is not important and task will be reused imedieately after previous task is  done
 
 ## WorkerTask
 Implementation fo the work executed by pool
 
 ### WorkStatus
    - Available
    - Working
    - Finished
    - Error

### Methods
    - IsAvailable() bool
    - IsFinished() bool
    - IsWorking() bool
    - IsError() bool
    - Result() (interface{}, error)
    - WaitFinished() (interface{}, error)
