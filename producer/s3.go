package producer

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/segmentio/ksuid"
)

type s3Producer struct {
	client *s3.S3
	bucket string

	buffer     []*Message
	timer      <-chan time.Time
	timerFired bool

	input    chan *Message
	inputWg  sync.WaitGroup
	output   chan Messages
	outputWg sync.WaitGroup
	errors   chan *Error

	FlushMessages  int
	FlushFrequency time.Duration
}

// NewS3Producer creates a producer that sends messages to AWS S3.
func NewS3Producer(endpoint, region, bucket string) (Producer, error) {
	// Configure to use Minio Server
	config := &aws.Config{
		Region:           aws.String(region),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}
	if endpoint != "" {
		config.Endpoint = aws.String(endpoint)
		config.DisableSSL = aws.Bool(!strings.Contains(endpoint, "https"))
	}

	sess, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}

	p := &s3Producer{
		client:         s3.New(sess),
		bucket:         bucket,
		input:          make(chan *Message),
		output:         make(chan Messages, 10),
		errors:         make(chan *Error),
		FlushMessages:  20000,
		FlushFrequency: 5 * time.Second,
	}

	go p.dispatchMessages()
	go p.dispatchFiles()

	return p, nil
}

// Name is the name of the producer.
func (p *s3Producer) Name() string {
	return "s3"
}

// Input is the message input channel.
func (p *s3Producer) Input() chan<- *Message {
	return p.input
}

// Errors is the error output channel.
func (p *s3Producer) Errors() <-chan *Error {
	return p.errors
}

// Close closes the producer.
func (p *s3Producer) Close() error {
	close(p.input)
	p.inputWg.Wait()

	// Flush last message set if it exists
	if len(p.buffer) > 0 {
		p.output <- p.buffer
	}

	// Wait for last messages to flush
	close(p.output)
	p.outputWg.Wait()

	close(p.errors)

	return nil
}

// IsHealthy checks the health of the producer.
func (p *s3Producer) IsHealthy() bool {
	return true
}

func (p *s3Producer) dispatchMessages() {
	p.inputWg.Add(1)
	defer p.inputWg.Done()

	if p.buffer == nil {
		p.buffer = newMessageBuffer(p.FlushMessages)
	}

	for {
		select {
		case msg, ok := <-p.input:
			if !ok {
				return
			}

			p.buffer = append(p.buffer, msg)

			if p.timer == nil {
				p.timer = time.After(p.FlushFrequency)
			}

		case <-p.timer:
			p.timerFired = true

		}

		if len(p.buffer) > p.FlushMessages || p.timerFired {
			p.output <- p.buffer
			p.buffer = newMessageBuffer(p.FlushMessages)
			p.timer = nil
			p.timerFired = false
		}
	}
}

func (p *s3Producer) dispatchFiles() {
	p.outputWg.Add(1)
	defer p.outputWg.Done()

	for msgs := range p.output {
		b, err := json.Marshal(&msgs)
		if err != nil {
			p.errors <- &Error{
				Msgs: msgs,
				Err: err,
			}
			continue
		}

		key := aws.String(ksuid.New().String() + ".json")

		_, err = p.client.PutObject(&s3.PutObjectInput{
			Body:   bytes.NewReader(b),
			Bucket: aws.String(p.bucket),
			Key:    key,
		})
		if err != nil {
			p.errors <- &Error{
				Msgs: msgs,
				Err: err,
			}
		}
	}
}

func newMessageBuffer(cap int) []*Message {
	return make([]*Message, 0, cap)
}
