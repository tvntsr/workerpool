package pool

import (
	"sync"
	"testing"
)

func Test_WorkerTask(t *testing.T) {
	var task Task

	task = &WorkerTask{
		lock:   sync.Mutex{},
		status: Available,
		id:     1,
		result: nil,
		error:  nil,
		comm:   make(chan func() (interface{}, error)),
		cond:   nil,
	}

	if !task.IsAvailable() {
		t.Fatal("Wrong task's status")
	}
}
