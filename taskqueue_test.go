package pool

import (
	"testing"
	"time"
)

func Test_WorkerTaskQueueAddAndRemoveTask(t *testing.T) {

	q := NewWorkerTaskQueue()

	if len(q.elements) != 0 {
		t.Fatal("Wrong elements count on new queue")
	}

	wt0, _ := createWorkerTask(0)
	wt1, _ := createWorkerTask(1)

	r := q.AddTask(wt0)
	if !r {
		t.Fatal("Expecting true in the response")
	}

	if len(q.elements) != 1 {
		t.Fatal("Wrong elements count on queue")
	}

	r = q.AddTask(wt1)
	if !r {
		t.Fatal("Expecting true in the response for second worker")
	}
	if len(q.elements) != 2 {
		t.Fatal("Wrong elements count on queue")
	}

	r = q.AddTask(wt0)
	if r {
		t.Fatal("Expecting false when adding self second time")
	}

	if len(q.elements) != 2 {
		t.Fatal("Wrong elements count on queue")
	}

	r = q.RemoveTask(wt0)
	if !r {
		t.Fatal("Expecting true when removing task")
	}
	if len(q.elements) != 1 {
		t.Fatal("Wrong elements count on queue")
	}

	r = q.RemoveTask(wt0)
	if r {
		t.Fatal("Expecting false when removing self second time")
	}
	if len(q.elements) != 1 {
		t.Fatal("Wrong elements count on queue")
	}

	r = q.RemoveTask(wt1)
	if !r {
		t.Fatal("Expecting true when removing self")
	}
	if len(q.elements) != 0 {
		t.Fatal("Wrong elements count on queue")
	}

}

func Test_TaimedWaitTaskFinishedTimeout(t *testing.T) {
	q := NewWorkerTaskQueue()
	wt0, _ := createWorkerTask(0)

	_ = q.AddTask(wt0)

	w, err := q.TaimedWaitTaskFinished(100 * time.Millisecond)
	if w != nil {
		t.Fatal("Wrong element returned")
	}
	if err == nil {
		t.Fatal("Expecting error")
	}
	if err.Error() != "timeout" {
		t.Fatal("Wrong error returned")
	}
}

func Test_TaimedWaitTaskFinished(t *testing.T) {
	q := NewWorkerTaskQueue()
	wt0, _ := createWorkerTask(0)
	wt1, _ := createWorkerTask(1)

	_ = q.AddTask(wt0)
	_ = q.AddTask(wt1)

	go func() {
		time.Sleep(1 * time.Second)
		wt0.setStatus(Finished, "", nil)
	}()

	w, err := q.TaimedWaitTaskFinished(3 * time.Second)
	if w == nil {
		t.Fatal("Wrong element returned")
	}
	if err != nil {
		t.Fatalf("Error returned, %v", err)
	}
	if w.id != wt0.id {
		t.Fatal("Wrong item returned")
	}
	if len(q.elements) != 1 {
		t.Fatal("Wrong elements count on queue")
	}

}

func Test_WaitTaskFinished(t *testing.T) {
	q := NewWorkerTaskQueue()
	wt0, _ := createWorkerTask(0)
	wt1, _ := createWorkerTask(1)

	_ = q.AddTask(wt0)
	_ = q.AddTask(wt1)

	go func() {
		time.Sleep(1 * time.Second)
		wt0.setStatus(Finished, "", nil)
	}()

	w, err := q.WaitTaskFinished()
	if w == nil {
		t.Fatal("Wrong element returned")
	}
	if err != nil {
		t.Fatalf("Error returned, %v", err)
	}
	if w.id != wt0.id {
		t.Fatal("Wrong item returned")
	}
	if len(q.elements) != 1 {
		t.Fatal("Wrong elements count on queue")
	}
}
