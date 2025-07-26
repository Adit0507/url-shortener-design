package snowflake

import (
	"sync"
	"time"
)

const (
	Epoch          int64 = 1288834974657
	WorkerIDBits         = 10
	SequenceBits         = 12
	MaxWorkerID          = 1 << WorkerIDBits
	MaxSequence          = 1 << SequenceBits
	TimestampShift       = WorkerIDBits + SequenceBits
	WorkerIDShift        = SequenceBits
)

type Snowflake struct {
	mu       sync.Mutex
	workerID int64
	sequence int64
	lastTime int64
}

func NewSnowFlake(workerID int64) *Snowflake {
	if workerID >= MaxWorkerID {
		panic("worker ID too large")
	}

	return &Snowflake{workerID: workerID}
}

func (s *Snowflake) Generate() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli() - Epoch
	if now == s.lastTime {
		s.sequence = (s.sequence + 1) & (MaxSequence - 1)
		if s.sequence == 0 {
			for now <= s.lastTime {
				now = time.Now().UnixMilli() - Epoch
				time.Sleep(time.Microsecond)
			}
		}
	} else {
		s.sequence = 0
	}

	s.lastTime = now
	id := (now << TimestampShift) | (s.workerID << WorkerIDShift) | s.sequence

	return id
}
