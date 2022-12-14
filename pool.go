package pool

/*
usage:
   working_pool := pool.NewPool(size of workers)

   working_pool.Start()

   status := working_pool.PushTask(func(){
   // something to run
   })

   status.IsFinished()
   status.IsError()
   status.IsStarted()
   status.IsWaiting()
   status.IsWorking()
*/

type Pool interface {
	Start() error
	Stop() error
	PushTask(func() (interface{}, error)) (Task, error)
	ReleaseTask(Task) error
	// RunUnmanaged(func() (interface{}, error)) error
	// WaitAllUnmanaged() error
	// WaitAllStarted() error
}
