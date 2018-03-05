package streaming

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
	"io/ioutil"
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
	}
	if endpoint != "" {
		config.Endpoint = aws.String(endpoint)
		config.DisableSSL = aws.Bool(!strings.Contains(endpoint, "https"))
		config.S3ForcePathStyle = aws.Bool(true)
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

		if len(p.buffer) >= p.FlushMessages || p.timerFired {
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

type s3Consumer struct {
	sess *session.Session
	client *s3.S3
	bucket string
}

// NewS3Consumer creates a consumer that gets messages to AWS S3.
func NewS3Consumer(endpoint, region, bucket string) (Consumer, error) {
	// Configure to use Minio Server
	config := &aws.Config{
		Region:           aws.String(region),
	}
	if endpoint != "" {
		config.Endpoint = aws.String(endpoint)
		config.DisableSSL = aws.Bool(!strings.Contains(endpoint, "https"))
		config.S3ForcePathStyle = aws.Bool(true)
	}

	sess, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}

	c := &s3Consumer{
		sess: sess,
		client: s3.New(sess),
		bucket: bucket,
	}

	return c, nil
}

// Output gets messages until the given date.
func (c *s3Consumer) Output(t time.Time) (<-chan Messages, <-chan error) {
	ch := make(chan Messages, 10)
	errors := make(chan error, 10)


	go func() {
		defer close(ch)
		defer close(errors)

		resp, err := c.client.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(c.bucket)})
		if err != nil {
			return
		}

		// loop bucket objects
		for _, item := range resp.Contents {
			if item.LastModified.After(t) {
				// since ksuid are sorted by timestamp, we can stop here
				break
			}
			object, err := c.client.GetObject(&s3.GetObjectInput{Bucket: aws.String(c.bucket), Key: item.Key})
			if err != nil {
				continue
			}

			msgs := Messages{}
			buf, err := ioutil.ReadAll(object.Body)
			if err != nil {
				errors <- err
				continue
			}

			err = json.Unmarshal(buf, &msgs)
			if err != nil {
				errors <- err
				continue
			}

			_, err = c.client.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(c.bucket), Key: item.Key})
			if err != nil {
				errors <- err
				continue
			}

			ch <- msgs
		}
	}()

	return ch, errors
}

func (c *s3Consumer) Close() error {
	return nil
}

func (c *s3Consumer) IsHealthy() bool {
	return true
}