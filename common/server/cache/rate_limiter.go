package cache

import (
	"sync"
	"time"
)

type ipRateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*rlBucket
	rate    float64 // tokens added per second
	burst   float64 // maximum tokens (and initial fill)
	lastGC  time.Time
}

type rlBucket struct {
	tokens float64
	last   time.Time
}
