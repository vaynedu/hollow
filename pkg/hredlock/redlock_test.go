package hredlock

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewRedlock(t *testing.T) {
	Convey("NewRedlock", t, func() {
		Convey("nil client 返回 ErrNilClient", func() {
			_, err := NewRedlock(nil)
			So(err, ShouldEqual, ErrNilClient)
		})

		Convey("非 nil client 构造成功", func() {
			client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
			defer client.Close()
			rl, err := NewRedlock(client)
			So(err, ShouldBeNil)
			So(rl, ShouldNotBeNil)
		})
	})
}

func TestLock_NilRedlock(t *testing.T) {
	Convey("nil Redlock.Lock 返回 ErrNilClient", t, func() {
		var rl *Redlock
		_, err := rl.Lock(context.Background(), "k", time.Second)
		So(err, ShouldEqual, ErrNilClient)
	})
}

func TestLock_UnreachableRedis(t *testing.T) {
	Convey("不可达 Redis 加锁失败", t, func() {
		client := redis.NewClient(&redis.Options{
			Addr:        "127.0.0.1:1",
			DialTimeout: 100 * time.Millisecond,
		})
		defer client.Close()
		rl, _ := NewRedlock(client)

		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		_, err := rl.Lock(ctx, "test-key", time.Second)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "lock")
	})
}

func TestUnlock_NilMutex(t *testing.T) {
	Convey("nil Mutex.Unlock", t, func() {
		Convey("nil receiver 返回 ErrNilMutex", func() {
			var m *Mutex
			err := m.Unlock(context.Background())
			So(err, ShouldEqual, ErrNilMutex)
		})
		Convey("nil 内部 m 返回 ErrNilMutex", func() {
			m := &Mutex{}
			err := m.Unlock(context.Background())
			So(err, ShouldEqual, ErrNilMutex)
		})
	})
}
