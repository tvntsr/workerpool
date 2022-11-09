package pool

import (
	"sync"
	"testing"
	"time"
)

func Test_NewScheduler(t *testing.T) {
	pool, err := NewManagedPool(5)
	if err != nil {
		t.Fatalf("Pool creation failed, %v", err)
	}

	sc, err := NewScheduler(pool)
	if err != nil {
		t.Fatalf("Pool creation failed, %v", err)
	}

	if sc == nil {
		t.Fatal("Scheduler is NIL!")
	}

}

func Test_SchedulerStop(t *testing.T) {
	pool, err := NewManagedPool(5)
	if err != nil {
		t.Fatalf("Pool creation failed, %v", err)
	}

	sc, err := NewScheduler(pool)
	if err != nil {
		t.Fatalf("Pool creation failed, %v", err)
	}

	status := sc.Stop()
	if !status {
		t.Error("Stop failed")
	}
}

func Test_SchedulerDoubleStop(t *testing.T) {
	pool, err := NewManagedPool(5)
	if err != nil {
		t.Fatalf("Pool creation failed, %v", err)
	}

	sc, err := NewScheduler(pool)
	if err != nil {
		t.Fatalf("Pool creation failed, %v", err)
	}

	status := sc.Stop()
	if !status {
		t.Fatal("Stop failed")
	}

	// second stop, must fail
	status = sc.Stop()
	if status {
		t.Fatal("Second stop() must fail")
	}
}

func Test_SchedulerRunTaskAt(t *testing.T) {
	pool, err := NewManagedPool(5)
	if err != nil {
		t.Fatalf("Pool creation failed, %v", err)
	}

	sc, err := NewScheduler(pool)
	if err != nil {
		t.Fatalf("Pool creation failed, %v", err)
	}
	err = pool.Start()
	if err != nil {
		t.Fatal(err)
	}

	a := 1
	r := &sync.WaitGroup{}
	r.Add(1)
	t.Log("Run method\n")
	task, err := sc.RunTaskAt(1*time.Second,
		func() (interface{}, error) {
			t.Log("I am started\n")

			a++
			r.Done()

			return nil, nil
		})

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	if task == nil {
		t.Error("Task should not be nil")
	}

	r.Wait()

	if a != 2 {
		t.Error("Task was not executed")
	}
	_ = sc.Stop()
	_ = pool.Stop()

}
