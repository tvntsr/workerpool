package pool

type Task interface {
	IsWaiting() bool
	IsFinished() bool
	IsStarted() bool
	IsWorking() bool
	IsError() bool
	Result() (interface{}, error)
	WaitFinished() (interface{}, error)
}
