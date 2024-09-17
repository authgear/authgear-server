package backoff

import (
	"math"
	"math/rand/v2"
	"time"
)

// Counter is a simple counter for exponential backoff mechanism with random jitter
type Counter struct {
	Interval    time.Duration
	MaxInterval time.Duration
	value       int
}

func (c *Counter) Increment() {
	c.value++
}

func (c *Counter) Reset() {
	c.value = 0
}

func (c *Counter) BackoffDuration() time.Duration {
	if c.value == 0 {
		return 0
	}

	duration := time.Duration(float64(c.Interval) * math.Pow(2, float64(c.value-1)))
	jitter := rand.N(time.Second)
	duration += jitter

	if c.MaxInterval != 0 {
		duration = min(duration, c.MaxInterval)
	}
	return duration
}
