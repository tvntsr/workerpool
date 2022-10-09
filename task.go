package pool

type Task interface {
	IsWaiting() bool
	IsFinished() bool
	IsWorking() bool
	IsError() bool
	Result() (interface{}, error)
	WaitFinished() (interface{}, error)
}
