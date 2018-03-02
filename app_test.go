package doubleteam_test

import (
	"context"
	"testing"
	"time"

	"github.com/msales/double-team"
	"github.com/msales/double-team/producer"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestSendsMessageToProducer(t *testing.T) {
	count := 0
	p := newFuncProducer(func(m *producer.Message) {
		count++
		assert.Equal(t, "test", m.Topic)
		assert.Equal(t, "test", string(m.Data))
	})
	app := doubleteam.NewApplication(context.Background(), []producer.Producer{p}, 1)
	defer app.Close()

	app.Send("test", []byte("test"))

	// Wait for the message to be processed
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 1, count)
}

func TestIsUnhealthyIfRecordsAreBlackHoled(t *testing.T) {
	p := newErrorProducer()
	app := doubleteam.NewApplication(context.Background(), []producer.Producer{p}, 1)
	defer app.Close()

	err := app.IsHealthy()
	assert.NoError(t, err)

	app.Send("test", []byte("test"))

	// Wait for the message to be processed
	time.Sleep(100 * time.Millisecond)

	err = app.IsHealthy()
	assert.Error(t, err)

	err = app.IsHealthy()
	assert.Error(t, err)
}

func TestCloseReturnsProducerErrors(t *testing.T) {
	p := newErrorProducer()
	app := doubleteam.NewApplication(context.Background(), []producer.Producer{p}, 1)

	err := app.Close()
	assert.Error(t, err)
}

type errorProducer struct {
	input  chan *producer.Message
	errors chan *producer.Error
}

func newErrorProducer() (producer.Producer) {
	p := &errorProducer{
		input:  make(chan *producer.Message),
		errors: make(chan *producer.Error),
	}

	go func() {
		for msg := range p.input {
			p.errors <- &producer.Error{
				Msgs: producer.Messages{msg},
				Err:  errors.New("test"),
			}
		}
	}()

	return p
}

func (p *errorProducer) Name() string {
	return "error-producer"
}

func (p *errorProducer) Input() chan<- *producer.Message {
	return p.input
}

func (p *errorProducer) Errors() <-chan *producer.Error {
	return p.errors
}

func (p *errorProducer) Close() error {
	close(p.input)
	close(p.errors)

	return errors.New("test")
}

func (p *errorProducer) IsHealthy() bool {
	return true
}

type funcProducer struct {
	input  chan *producer.Message
	errors chan *producer.Error
}

func newFuncProducer(fn func(message *producer.Message)) (producer.Producer) {
	p := &errorProducer{
		input:  make(chan *producer.Message),
		errors: make(chan *producer.Error),
	}

	go func() {
		for msg := range p.input {
			fn(msg)
		}
	}()

	return p
}

func (p *funcProducer) Name() string {
	return "error-producer"
}

func (p *funcProducer) Input() chan<- *producer.Message {
	return p.input
}

func (p *funcProducer) Errors() <-chan *producer.Error {
	return p.errors
}

func (p *funcProducer) Close() error {
	close(p.input)
	close(p.errors)

	return nil
}

func (p *funcProducer) IsHealthy() bool {
	return true
}
