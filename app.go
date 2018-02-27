package double_team

import (
	"context"
	"errors"
	"fmt"

	"github.com/msales/double-team/producer"
	"github.com/msales/pkg/log"
	"github.com/msales/pkg/stats"
)

type applicationErrors []error

func (ae applicationErrors) Error() string {
	return fmt.Sprintf("app: Failed to close %d producers.", len(ae))
}

// Application represents the application.
type Application struct {
	producers []producer.Producer
	messages  chan *producer.Message
}

// NewApplication creates an instance of Application.
func NewApplication(ctx context.Context, producers []producer.Producer) *Application {
	app := &Application{
		producers: producers,
		messages:  make(chan *producer.Message, 1000),
	}

	// Wire the producer chain
	ch := &app.messages
	for _, p := range app.producers {
		go func(ch *chan *producer.Message, p producer.Producer) {
			for msg := range *ch {
				p.Input() <- msg
				stats.Inc(ctx, "produced", 1, 1.0, map[string]string{"queue": p.Name()})
			}
		}(ch, p)

		newCh := make(chan *producer.Message, 1000)
		ch = &newCh
		go func(ch *chan *producer.Message, p producer.Producer) {
			for err := range p.Errors() {
				for _, msg := range err.Msgs {
					*ch <- msg
					stats.Inc(ctx, "error", 1, 1.0, map[string]string{"queue": p.Name()})
				}
				log.Error(ctx, "app: error producing message", "queue", p.Name(), "error", err.Err.Error())
			}
			close(*ch)
		}(ch, p)
	}

	// Wire the black-hole
	go func(ch *chan *producer.Message) {
		for _ = range *ch {
			stats.Inc(ctx, "produced", 1, 1.0, map[string]string{"queue": "black-hole"})
		}
	}(ch)

	return app
}

// Send sends a message to the producer chain.
func (a *Application) Send(topic string, data []byte) {
	a.messages <- &producer.Message{
		Topic: topic,
		Data:  data,
	}
}

// Close closes the application and cleans up.
func (a *Application) Close() error {
	close(a.messages)

	var errs applicationErrors
	for _, p := range a.producers {
		if err := p.Close(); err != nil {
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
	for _, p := range a.producers {
		if ok := p.IsHealthy(); !ok {
			return errors.New("unhealthy producer")
		}
	}

	return nil
}
