package doubleteam

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/msales/double-team/streaming"
	"github.com/msales/pkg/stats"
	"github.com/pkg/errors"
)

type applicationErrors []error

func (ae applicationErrors) Error() string {
	return fmt.Sprintf("app: Failed to close %d producers cleanly.", len(ae))
}

var errUnhealthy = errors.New("app: service unhealthy")

// Application represents the application.
type Application struct {
	producers []streaming.Producer
	messages  chan *streaming.Message

	statsTimer *time.Ticker

	errorCount  int64
	unhealthy   bool
	closeErrors chan error
}

// NewApplication creates an instance of Application.
func NewApplication(ctx context.Context, producers []streaming.Producer, queueSize int) *Application {
	closeMutex := sync.WaitGroup{}
	app := &Application{
		producers:   producers,
		messages:    make(chan *streaming.Message, queueSize),
		closeErrors: make(chan error),
	}

	channels := map[string]*chan *streaming.Message{}

	// Wire the producer chain
	ch := &app.messages
	for _, p := range app.producers {
		channels[p.Name()] = ch
		go func(ch *chan *streaming.Message, p streaming.Producer) {
			for msg := range *ch {
				p.Input() <- msg
				stats.Inc(ctx, "produced", 1, 1.0, map[string]string{"queue": p.Name()})
			}

			closeMutex.Add(1)
			app.closeErrors <- p.Close()
			closeMutex.Done()
		}(ch, p)

		newCh := make(chan *streaming.Message, queueSize)
		ch = &newCh
		go func(ch *chan *streaming.Message, p streaming.Producer) {
			for err := range p.Errors() {
				for _, msg := range err.Msgs {
					*ch <- msg
					stats.Inc(ctx, "error", 1, 1.0, map[string]string{"queue": p.Name()})
				}
			}
			close(*ch)
		}(ch, p)
	}

	// Wire the black-hole
	go func(ch *chan *streaming.Message) {
		for range *ch {
			atomic.AddInt64(&app.errorCount, 1)
			stats.Inc(ctx, "produced", 1, 1.0, map[string]string{"queue": "black-hole"})
		}

		closeMutex.Wait()
		close(app.closeErrors)
	}(ch)

	app.statsTimer = time.NewTicker(1 * time.Second)
	go func() {
		for range app.statsTimer.C {
			for k, ch := range channels {
				stats.Gauge(ctx, "queue_length", float64(len(*ch)), 1.0, map[string]string{"queue": k})
				stats.Gauge(ctx, "queue_length_max", float64(cap(*ch)), 1.0, map[string]string{"queue": k})
			}
		}
	}()

	return app
}

// Send sends a message to the producer chain.
func (a *Application) Send(topic string, key, data []byte) {
	a.messages <- &streaming.Message{
		Topic: topic,
		Key:   key,
		Data:  data,
	}
}

// Close closes the application and cleans up.
func (a *Application) Close() error {
	close(a.messages)

	var errs applicationErrors
	for err := range a.closeErrors {
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// IsHealthy checks the health of the Application.
func (a *Application) IsHealthy() error {
	if a.unhealthy {
		return errUnhealthy
	}

	errs := atomic.LoadInt64(&a.errorCount)
	if errs > 0 {
		a.unhealthy = true
		return errUnhealthy
	}

	return nil
}
