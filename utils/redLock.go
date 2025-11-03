package utils

import (
	"context"
	"errors"
	"github.com/go-redsync/redsync/v4"
	redsyncredis "github.com/go-redsync/redsync/v4/redis"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedLockCreater struct {
	RS *redsync.Redsync
}

func NewRedLockCreater(clis []*redis.Client) (*RedLockCreater, error) {
	if len(clis) == 0 {
		return nil, errors.New("redis clients is nil")
	}

	var pools []redsyncredis.Pool

	for _, cli := range clis {
		p := goredis.NewPool(cli)
		pools = append(pools, p)
	}

	rs := redsync.New(pools...)

	return &RedLockCreater{
		RS: rs,
	}, nil
}

func (r *RedLockCreater) GetLock(ctx context.Context, lockName string, retryCount redsync.Option) (*redsync.Mutex, error) {
	//redsync.WithTries(32)
	m := r.RS.NewMutex(lockName, redsync.WithExpiry(30*time.Second), retryCount)
	e := m.Lock()
	if e != nil {
		return nil, e
	}
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				ok, err := m.Extend()
				if err != nil || !ok {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return m, nil
}

func (r *RedLockCreater) ReleaseLock(m *redsync.Mutex, cancel context.CancelFunc) error {
	cancel()
	ok, err := m.Unlock()
	if err != nil || !ok {
		return err
	}
	return nil
}
