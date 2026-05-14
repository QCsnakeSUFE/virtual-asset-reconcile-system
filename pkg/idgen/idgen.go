package idgen

import (
	"sync"
	"time"
)

const (
	workerBits  = 10 // 机器/节点编号
	seqBits     = 12 // 同一毫秒内的自增序号
	workerShift = seqBits
	timeShift   = seqBits + workerBits
	seqMask     = 1<<seqBits - 1 //序列号的掩码（最大值）
	epoch       = 1700000000000  // 自定义纪元（毫秒级）
)

type Snowflake struct {
	mu       sync.Mutex // 线程安全锁
	workerID int64
	lastTS   int64
	sequence int64
}

func New(workerID int64) *Snowflake {
	return &Snowflake{workerID: workerID}
}

func (s *Snowflake) NextID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()
	if now == s.lastTS {
		// 同一毫秒内，继续生成 ID
		s.sequence = (s.sequence + 1) & seqMask
		if s.sequence == 0 {
			// 若序列号用完，则需要等下一毫秒
			for now <= s.lastTS {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		// 新的一毫秒，序号重置为 0
		s.sequence = 0
	}
	s.lastTS = now
	return (now-epoch)<<timeShift | (s.workerID << workerShift) | s.sequence
}

var defaultGen = New(0)

func NextID() int64 {
	return defaultGen.NextID()
}
