package breaker

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// ErrBreakerOpen is the error returned from Run() when the breaker is open.
var ErrBreakerOpen = errors.New("breaker: circuit breaker is open")

const (
	closed uint32 = iota
	open
)

// Breaker implements the circuit-breaker pattern.
type Breaker struct {
	errorThreshold int
	timeout        time.Duration

	lock      sync.Mutex
	state     uint32
	errors    int
	lastError time.Time
}

// New creates a new Breaker.
func New(errorThreshold int, timeout time.Duration) *Breaker {
	return &Breaker{
		errorThreshold: errorThreshold,
		timeout:        timeout,
	}
}

// Run will run the given function or return ErrBreakerOpen immediately
// if the circuit-breaker is open.
func (b *Breaker) Run(fn func()) error {
	state := atomic.LoadUint32(&b.state)

	if state == open {
		return ErrBreakerOpen
	}

	fn()
	return nil
}

// Error registers an error with the circuit-breaker, potentially opening the breaker.
func (b *Breaker) Error() {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.errors > 0 {
		expiry := b.lastError.Add(b.timeout)
		if time.Now().After(expiry) {
			b.errors = 0
		}
	}

	if b.state == closed {
		b.errors++
		if b.errors == b.errorThreshold {
			b.openBreaker()
		} else {
			b.lastError = time.Now()
		}
	}
}

func (b *Breaker) openBreaker() {
	b.changeState(open)
	go b.timer()
}

func (b *Breaker) timer() {
	time.Sleep(b.timeout)

	b.lock.Lock()
	defer b.lock.Unlock()

	b.changeState(closed)
}

func (b *Breaker) changeState(newState uint32) {
	b.errors = 0
	atomic.StoreUint32(&b.state, newState)
}
