package router

import (
	"sync"
	"time"
)

type CircuitBreaker struct {
	failures     int
	maxFailures  int
	openUntil    time.Time
	openDuration time.Duration
	mu           sync.Mutex
}

func NewCircuitBreaker(maxFailures int, openDuration time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		openDuration: openDuration,
	}
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if time.Now().Before(cb.openUntil) {
		return false
	}
	return true
}

func (cb *CircuitBreaker) Success() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
}

func (cb *CircuitBreaker) Failure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	if cb.failures >= cb.maxFailures {
		cb.openUntil = time.Now().Add(cb.openDuration)
		cb.failures = 0
	}
}
