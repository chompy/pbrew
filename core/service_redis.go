package core

import (
	"bytes"
	"fmt"
	"strings"
)

// IsRedis returns true if service is redis.
func (s *Service) IsRedis() bool {
	return strings.HasPrefix(s.BrewAppName(), "redis")
}

// IsRedisRunning returns true if redis is running.
func (s *Service) IsRedisRunning() bool {
	c := NewShellCommand()
	c.Command = "pgrep"
	p, err := s.Port()
	if err != nil {
		return false
	}
	c.Args = []string{"-f", fmt.Sprintf("redis-server 127.0.0.1:%d", p)}
	var buf bytes.Buffer
	c.Stdout = &buf
	c.Interactive()
	return strings.TrimSpace(buf.String()) != ""
}
