package pool

import (
	"testing"
)

func Test_NewWorkerPool(t *testing.T) {

	pool_size := 10

	pool, err := NewWorkerPool(pool_size)
	if err != nil {
		t.Fatalf("Error %v", err)
	}

	if pool.max_size != pool_size {
		t.Fatal("Wrong pool size")
	}

	if pool.tasks != nil {
		t.Fatal("Wrong pool tasks list")
	}

	if pool.available != nil {
		t.Fatal("Wrong pool available list")
	}
}

func Test_WorkerPoolStart(t *testing.T) {

	pool_size := 10

	pool, err := NewWorkerPool(pool_size)
	if err != nil {
		t.Fatalf("Pool creating error %v", err)
	}

	err = pool.Start()
	if err != nil {
		t.Fatalf("Pool start error %v", err)
	}

	if len(pool.tasks) != pool_size {
		t.Fatal("Wrong tasks available list")
	}

	if len(pool.available) != pool_size {
		t.Fatal("Wrong pool available list")
	}
}

func Test_WorkerPoolStop(t *testing.T) {

	pool_size := 10

	pool, err := NewWorkerPool(pool_size)
	if err != nil {
		t.Fatalf("Pool creating error %v", err)
	}

	err = pool.Start()
	if err != nil {
		t.Fatalf("Pool start error %v", err)
	}

	err = pool.Stop()
	if err != nil {
		t.Fatalf("Pool stop error %v", err)
	}

	if len(pool.tasks) != 0 {
		t.Fatalf("Wrong tasks available list, %d", len(pool.tasks))
	}

	if len(pool.available) != 0 {
		t.Fatalf("Wrong pool available list, %d", len(pool.available))
	}
}
