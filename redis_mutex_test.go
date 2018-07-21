package common

import (
	"fmt"
	"github.com/go-redis/redis"
	"sync"
	"testing"
)

func TestMutex1(t *testing.T) {
	rds := redis.NewClient(&redis.Options{
		Addr:     "192.168.1.244:6379",
		Password: "",
		PoolSize: 10,
	})
	lock := NewRedisLock("test1")

	n := 0
	w := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		w.Add(1)
		go func() {
			defer w.Done()
			if err := lock.Lock(rds); err != nil {
				fmt.Println(err)
				return
			}
			n += 1
			fmt.Println(n)
			lock.UnLock(rds)
		}()
	}
	w.Wait()
	fmt.Println(n)
}

func TestMutex2(t *testing.T) {
	rds1 := redis.NewClient(&redis.Options{
		Addr:     "192.168.1.244:6379",
		Password: "",
		PoolSize: 10,
	})
	rds2 := redis.NewClient(&redis.Options{
		Addr:     "192.168.1.244:6379",
		Password: "",
		PoolSize: 10,
	})
	lock1 := NewRedisLock("test2")
	lock2 := NewRedisLock("test2")

	n := 0
	w := sync.WaitGroup{}
	for i := 0; i < 500; i++ {
		w.Add(1)
		go func() {
			defer w.Done()
			if err := lock1.Lock(rds1); err != nil {
				fmt.Println(err)
				return
			}
			n += 1
			fmt.Println(n)
			lock1.UnLock(rds1)
		}()
	}

	for i := 0; i < 500; i++ {
		w.Add(1)
		go func() {
			defer w.Done()
			if err := lock2.Lock(rds2); err != nil {
				fmt.Println(err)
				return
			}
			n += 1
			fmt.Println(n)
			lock2.UnLock(rds2)
		}()
	}

	w.Wait()
	fmt.Println(n)
}
