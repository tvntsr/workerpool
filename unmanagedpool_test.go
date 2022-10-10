package pool

import (
	"testing"
	"time"
)

func Test_NewUnmanagedPool(t *testing.T) {
	pool_size := 10

	var pool Pool
	pool, err := NewUnmanagedPool(pool_size)
	if err != nil {
		t.Fatalf("Error %v", err)
	}

	wpool := pool.(*UnmanagedPool)

	if wpool.max_size != pool_size {
		t.Fatal("Wrong pool size")
	}

	if wpool.tasks != nil {
		t.Fatal("Wrong pool tasks list")
	}
}

func Test_UnmanagedPoolStart(t *testing.T) {
	pool_size := 10

	pool, err := NewUnmanagedPool(pool_size)
	if err != nil {
		t.Fatalf("Pool creating error %v", err)
	}

	err = pool.Start()
	if err != nil {
		t.Fatalf("Pool start error %v", err)
	}

	if len(pool.tasks) != pool_size {
		t.Fatal("Wrong tasks list")
	}

	if len(pool.unmanaged) != 0 {
		t.Fatal("Wrong pool internal list")
	}
}

func Test_UnmanagedPoolStop(t *testing.T) {
	pool_size := 10

	pool, err := NewUnmanagedPool(pool_size)
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
		t.Fatalf("Wrong tasks list, %d", len(pool.tasks))
	}
}

func Test_UnmanagedPoolPushTask(t *testing.T) {
	pool_size := 2

	pool, err := NewUnmanagedPool(pool_size)
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
	if worker != nil {
		t.Fatal("Worker should be nil!")
		t.FailNow()
	}
}

func Test_UnmanagedPoolPushLongTask(t *testing.T) {
	pool_size := 2

	pool, err := NewUnmanagedPool(pool_size)
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
	if worker != nil {
		t.Fatal("Worker should be nil!")
		t.FailNow()
	}
}

func Test_UnmanagedPoolPushTooMuchTasks(t *testing.T) {
	pool_size := 2

	pool, err := NewUnmanagedPool(pool_size)
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

		if err != nil {
			t.Fatalf("Pool push task error %v", err)
			t.FailNow()
		}
		if worker != nil {
			t.Fatal("Worker should be nil!")
			t.FailNow()
		}
	}
}
