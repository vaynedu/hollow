package hidgenerator

import (
	"errors"
	"strconv"
	"sync"
	"time"
)

// Snowflake 64bit Twitter 标准实现:1 符号 + 41 毫秒时间戳 + 10 机器 ID + 12 序列
//
// 自定义 epoch:2020-01-01 UTC,可用到 ~2089 年
// 单机器单毫秒最多 4096 个 ID,理论峰值 ~4M/s
// 时钟回拨场景:简化处理,沿用上次时间戳防止 ID 倒退(可能导致序列号溢出 hang)
type Snowflake struct {
	mu        sync.Mutex
	epoch     int64 // 起始时间戳(毫秒)
	machineID int64 // 机器 ID,bit 21-12
	lastTime  int64 // 上次生成的毫秒时间戳
	sequence  int64 // 同毫秒内序列号
}

const (
	snowflakeEpoch       int64 = 1577836800000 // 2020-01-01 00:00:00 UTC ms
	snowflakeMachineBits uint  = 10
	snowflakeSeqBits     uint  = 12
	snowflakeMaxMachine  int64 = (1 << snowflakeMachineBits) - 1
	snowflakeMaxSeq      int64 = (1 << snowflakeSeqBits) - 1
)

// ErrMachineIDOutOfRange machineID 超出 [0, 1024) 范围时返回
var ErrMachineIDOutOfRange = errors.New("hidgenerator: machineID must be in [0, 1024)")

// NewSnowflake 构造 Snowflake 实例
// machineID 取值 [0, 1023],由部署方在集群内分配唯一值
func NewSnowflake(machineID int64) (*Snowflake, error) {
	if machineID < 0 || machineID > snowflakeMaxMachine {
		return nil, ErrMachineIDOutOfRange
	}
	return &Snowflake{
		epoch:     snowflakeEpoch,
		machineID: machineID,
	}, nil
}

// GenerateRequestID 实现 IdGenerator 接口,返回 base10 字符串形式的 snowflake ID
func (s *Snowflake) GenerateRequestID() string {
	return strconv.FormatInt(s.GenerateInt64(), 10)
}

// GenerateInt64 直接生成 int64 形式的 snowflake ID,避免字符串分配
func (s *Snowflake) GenerateInt64() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()
	if now < s.lastTime {
		// 时钟回拨,等到追回上次时间戳防止 ID 倒退
		now = s.lastTime
	}
	if now == s.lastTime {
		s.sequence = (s.sequence + 1) & snowflakeMaxSeq
		if s.sequence == 0 {
			// 序列溢出,等待下一毫秒
			for now <= s.lastTime {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		s.sequence = 0
	}
	s.lastTime = now

	elapsed := now - s.epoch
	return (elapsed << (snowflakeMachineBits + snowflakeSeqBits)) |
		(s.machineID << snowflakeSeqBits) |
		s.sequence
}
