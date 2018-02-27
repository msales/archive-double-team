package producer

import (
	"sync"
	"time"

	"github.com/Shopify/sarama"
)

type kafkaProducer struct {
	client   sarama.Client
	producer sarama.AsyncProducer

	input  chan *Message
	errors chan *Error
	wg     sync.WaitGroup
}

func NewKafkaProducer(brokers []string, retry int) (Producer, error) {
	config := sarama.NewConfig()
	config.Metadata.RefreshFrequency = 30 * time.Second
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Flush.Frequency = 100 * time.Millisecond
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
		input:    make(chan *Message),
		errors:   make(chan *Error),
	}

	go p.dispatchMessages()
	go p.dispatchErrors()

	return p, nil
}

func (p *kafkaProducer) Name() string {
	return "kafka"
}

func (p *kafkaProducer) Input() chan<- *Message {
	return p.input
}

func (p *kafkaProducer) Errors() <-chan *Error {
	return p.errors
}

func (p *kafkaProducer) Close() error {
	p.producer.AsyncClose()

	p.wg.Wait()

	close(p.input)
	close(p.errors)

	return p.client.Close()
}

// IsHealthy checks the health of the Kafka cluster.
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
		p.producer.Input() <- &sarama.ProducerMessage{
			Topic: msg.Topic,
			Value: sarama.ByteEncoder(msg.Data),
		}
	}
}

func (p *kafkaProducer) dispatchErrors() {
	p.wg.Add(1)

	for err := range p.producer.Errors() {
		p.errors <- &Error{
			Msgs: []*Message{
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
