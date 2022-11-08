package pool

import (
	"testing"
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
