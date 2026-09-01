package hidgenerator

import (
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewSnowflake(t *testing.T) {
	Convey("NewSnowflake", t, func() {
		Convey("machineID 在范围内构造成功", func() {
			s, err := NewSnowflake(0)
			So(err, ShouldBeNil)
			So(s, ShouldNotBeNil)

			s2, err2 := NewSnowflake(1023)
			So(err2, ShouldBeNil)
			So(s2, ShouldNotBeNil)
		})

		Convey("machineID 负数返回 ErrMachineIDOutOfRange", func() {
			_, err := NewSnowflake(-1)
			So(err, ShouldEqual, ErrMachineIDOutOfRange)
		})

		Convey("machineID >= 1024 返回 ErrMachineIDOutOfRange", func() {
			_, err := NewSnowflake(1024)
			So(err, ShouldEqual, ErrMachineIDOutOfRange)
		})
	})
}

func TestSnowflakeGenerateInt64(t *testing.T) {
	Convey("Snowflake.GenerateInt64", t, func() {
		s, _ := NewSnowflake(1)

		Convey("ID 单调递增", func() {
			a := s.GenerateInt64()
			b := s.GenerateInt64()
			c := s.GenerateInt64()
			So(a, ShouldBeLessThan, b)
			So(b, ShouldBeLessThan, c)
		})

		Convey("ID 永远大于 0", func() {
			for i := 0; i < 100; i++ {
				So(s.GenerateInt64(), ShouldBeGreaterThan, 0)
			}
		})
	})
}

func TestSnowflakeGenerateRequestID(t *testing.T) {
	Convey("Snowflake.GenerateRequestID 实现 IdGenerator", t, func() {
		var gen IdGenerator
		gen, err := NewSnowflake(1)
		So(err, ShouldBeNil)

		id1 := gen.GenerateRequestID()
		id2 := gen.GenerateRequestID()
		So(id1, ShouldNotBeEmpty)
		So(id2, ShouldNotBeEmpty)
		So(id1, ShouldNotEqual, id2)
	})
}

func TestSnowflakeConcurrent(t *testing.T) {
	Convey("并发生成 ID 无重复", t, func() {
		s, _ := NewSnowflake(1)
		const goroutines = 10
		const perGoroutine = 1000

		var mu sync.Mutex
		ids := make(map[int64]struct{}, goroutines*perGoroutine)

		var wg sync.WaitGroup
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				local := make([]int64, 0, perGoroutine)
				for j := 0; j < perGoroutine; j++ {
					local = append(local, s.GenerateInt64())
				}
				mu.Lock()
				for _, id := range local {
					ids[id] = struct{}{}
				}
				mu.Unlock()
			}()
		}
		wg.Wait()

		So(len(ids), ShouldEqual, goroutines*perGoroutine)
	})
}
