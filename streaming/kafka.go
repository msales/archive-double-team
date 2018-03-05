package streaming

import (
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/msales/double-team/pkg/breaker"
)

type kafkaProducer struct {
	client   sarama.Client
	producer sarama.AsyncProducer

	breaker *breaker.Breaker
	input   chan *Message
	errors  chan *Error
	wg      sync.WaitGroup
}

// NewKafkaProducer creates a new producer that sends messages to Kafka.
func NewKafkaProducer(brokers []string, retry int) (Producer, error) {
	config := sarama.NewConfig()
	config.Metadata.RefreshFrequency = 30 * time.Second
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Flush.Frequency = 500 * time.Millisecond
	config.Producer.Retry.Max = retry
	config.Producer.Retry.Backoff = 10 * time.Millisecond

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		return nil, err
	}

	producer, err := sarama.NewAsyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}

	p := &kafkaProducer{
		client:   client,
		producer: producer,
		breaker:  breaker.New(5, 1*time.Second),
		input:    make(chan *Message),
		errors:   make(chan *Error, 100),
	}

	go p.dispatchMessages()
	go p.dispatchErrors()

	return p, nil
}

// Name is the name of the producer.
func (p *kafkaProducer) Name() string {
	return "kafka"
}

// Input is the message input channel.
func (p *kafkaProducer) Input() chan<- *Message {
	return p.input
}

// Errors is the error output channel.
func (p *kafkaProducer) Errors() <-chan *Error {
	return p.errors
}

// Close closes the producer.
func (p *kafkaProducer) Close() error {
	close(p.input)

	p.producer.AsyncClose()
	p.wg.Wait()

	err := p.client.Close()

	close(p.errors)

	return err
}

// IsHealthy checks the health of the producer.
func (p *kafkaProducer) IsHealthy() bool {
	for _, b := range p.client.Brokers() {
		if ok, err := b.Connected(); ok || err == nil {
			return true
		}
	}

	return false
}

func (p *kafkaProducer) dispatchMessages() {
	for msg := range p.input {
		err := p.breaker.Run(func() {
			p.producer.Input() <- &sarama.ProducerMessage{
				Topic: msg.Topic,
				Value: sarama.ByteEncoder(msg.Data),
			}
		})

		if err != nil {
			p.errors <- &Error{
				Msgs: Messages{msg},
				Err:  err,
			}
		}
	}
}

func (p *kafkaProducer) dispatchErrors() {
	p.wg.Add(1)

	for err := range p.producer.Errors() {
		p.breaker.Error()
		p.errors <- &Error{
			Msgs: Messages{
				{
					Topic: err.Msg.Topic,
					Data:  err.Msg.Value.(sarama.ByteEncoder),
				},
			},
			Err: err.Err,
		}
	}

	p.wg.Done()
}
