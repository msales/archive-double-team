package breaker_test

import (
	"testing"
	"time"

	"github.com/msales/double-team/pkg/breaker"
	"github.com/stretchr/testify/assert"
)

func TestBreakerStartsClosed(t *testing.T) {
	b := breaker.New(2, 100*time.Millisecond)

	err := b.Run(func() {})
	assert.NoError(t, err)
}

func TestBreakerErrorExpiry(t *testing.T) {
	b := breaker.New(2, 100*time.Millisecond)

	for i := 0; i < 3; i++ {
		b.Error()
		time.Sleep(100 * time.Millisecond)
	}

	err := b.Run(func() {})
	assert.NoError(t, err)
}

func TestBreakerStateTransitions(t *testing.T) {
	b := breaker.New(2, 100*time.Millisecond)

	// Breaker is closed
	for i := 0; i < 3; i++ {
		b.Error()
	}

	// Breaker is open
	err := b.Run(func() {})
	assert.Equal(t, breaker.ErrBreakerOpen, err)

	// Wait for breaker to close
	time.Sleep(200 * time.Millisecond)

	// Breaker should now be closed
	err = b.Run(func() {})
	assert.NoError(t, err)
}
