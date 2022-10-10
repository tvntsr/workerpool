package pool

import (
	"testing"
	"time"
)

func Test_NewWorkerPool(t *testing.T) {
	pool_size := 10

	var pool Pool
	pool, err := NewWorkerPool(pool_size)
	if err != nil {
		t.Fatalf("Error %v", err)
	}

	wpool := pool.(*WorkerPool)

	if wpool.max_size != pool_size {
		t.Fatal("Wrong pool size")
	}

	if wpool.tasks != nil {
		t.Fatal("Wrong pool tasks list")
	}

	if wpool.available != nil {
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

func Test_WorkerPoolPushTask(t *testing.T) {
	pool_size := 2

	pool, err := NewWorkerPool(pool_size)
	if err != nil {
		t.Fatalf("Pool creating error %v", err)
	}

	err = pool.Start()
	if err != nil {
		t.Fatalf("Pool start error %v", err)
	}

	worker, err := pool.PushTask(func() (interface{}, error) {
		return 1, nil
	})

	if err != nil {
		t.Fatalf("Pool push task error %v", err)
	}
	if worker == nil {
		t.Fatal("Worker is nil!")
		t.FailNow()
	}

	ret, err := worker.WaitFinished()
	if err != nil {
		t.Fatalf("WaitFinished failed %v", err)
	}
	if ret != 1 {
		t.Fatalf("Wrong value returned, %v", ret)
	}
}

func Test_WorkerPoolPushLongTask(t *testing.T) {
	pool_size := 2

	pool, err := NewWorkerPool(pool_size)
	if err != nil {
		t.Fatalf("Pool creating error %v", err)
	}

	err = pool.Start()
	if err != nil {
		t.Fatalf("Pool start error %v", err)
	}

	worker, err := pool.PushTask(func() (interface{}, error) {
		time.Sleep(3 * time.Second)
		return 1, nil
	})

	if err != nil {
		t.Fatalf("Pool push task error %v", err)
	}
	if worker == nil {
		t.Fatal("Worker is nil!")
		t.FailNow()
	}

	ret, err := worker.WaitFinished()
	if err != nil {
		t.Fatalf("WaitFinished failed %v", err)
	}
	if ret != 1 {
		t.Fatalf("Wrong value returned, %v", ret)
	}
}

func Test_WorkerPoolPushTooMuchTasks(t *testing.T) {
	pool_size := 2

	pool, err := NewWorkerPool(pool_size)
	if err != nil {
		t.Fatalf("Pool creating error %v", err)
	}

	err = pool.Start()
	if err != nil {
		t.Fatalf("Pool start error %v", err)
	}

	for i := 0; i < pool_size+2; i++ {
		worker, err := pool.PushTask(func() (interface{}, error) {
			time.Sleep(3 * time.Second)
			return 1, nil
		})

		if err == nil && i > pool_size {
			t.Fatalf("Pool pushed too much tasks, %d", i)
		}

		if err != nil && i < pool_size {
			t.Fatalf("Pool push task error %v", err)
			t.FailNow()
		}
		if worker == nil && i < pool_size {
			t.Fatal("Worker is nil!")
			t.FailNow()
		}
	}
}
