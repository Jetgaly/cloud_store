package utils

import (
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/snowflake"
)

type SafeSnowFlakeCreater struct {
	node        *snowflake.Node
	maxWaitTime time.Duration
	lastTime    int64
	mu          sync.Mutex
}

func NewSafeSnowFlakeCreater(nodeId int64, maxWaitTime time.Duration) (*SafeSnowFlakeCreater, error) {
	node, err := snowflake.NewNode(nodeId)
	if err != nil {
		return nil, err
	}

	return &SafeSnowFlakeCreater{
		node:        node,
		maxWaitTime: maxWaitTime,
	}, nil
}

func (s *SafeSnowFlakeCreater) Generate() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for {
		id := s.node.Generate()
		currentTime := int64(id) >> (snowflake.NodeBits + snowflake.StepBits)

		if s.lastTime == 0 {
			s.lastTime = currentTime
			return int64(id), nil
		}

		if currentTime < s.lastTime {
			timeDiff := s.lastTime - currentTime
			waitDuration := time.Duration(timeDiff) * time.Millisecond

			if waitDuration <= s.maxWaitTime {
				time.Sleep(waitDuration)
				continue
			} else {
				return 0, fmt.Errorf("时钟回拨过大: %dms，超过最大等待时间", timeDiff)
			}
		}

		s.lastTime = currentTime
		return int64(id), nil
	}
}
