package pool

type Task interface {
	IsAvailable() bool
	IsFinished() bool
	IsWorking() bool
	IsError() bool
	Result() (interface{}, error)
	WaitFinished() (interface{}, error)
}
