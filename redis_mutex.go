package redislock

import (
	"encoding/base64"
	"errors"
	"github.com/go-redis/redis"
	"math/rand"
	"sync"
	"time"
)

const (
	DefaultRetry    = 5
	DefaultInterval = 100 * time.Millisecond
	DefaultExpire   = 5 * time.Second
)

var ErrFailed = errors.New("failed to acquire lock")

type RedisLock struct {
	Expiry        time.Duration
	Retry         int
	RetryInterval time.Duration

	lockKey   string
	lockValue string
	mutex     sync.Mutex
}

func NewRedisLock(lk string) *RedisLock {
	return &RedisLock{lockKey: lk, Expiry: DefaultExpire, Retry: DefaultRetry, RetryInterval: DefaultInterval}
}

func (rl *RedisLock) Lock(rds *redis.Client) error {
	rl.mutex.Lock()
	err := rl.lock(rds)
	if err != nil {
		rl.mutex.Unlock()
	}
	return err
}

func (rl *RedisLock) lock(rds *redis.Client) error {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return err
	}
	rl.lockValue = base64.StdEncoding.EncodeToString(b)
	for i := 0; i < rl.Retry; i++ {
		ok, err := rds.SetNX(rl.lockKey, rl.lockValue, rl.Expiry).Result()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		if i < rl.Retry-1 {
			time.Sleep(rl.RetryInterval)
		}
	}
	return ErrFailed
}

const delScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
else
	return 0
end`

func (rl *RedisLock) UnLock(rds *redis.Client) {
	rds.Eval(delScript, []string{rl.lockKey}, rl.lockValue).Result()
	rl.mutex.Unlock()
}
