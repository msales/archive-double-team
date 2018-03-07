package streaming

import "time"

// Messages is an array of messages.
type Messages []*Message

// Message is the information to be sent through a Producer.
type Message struct {
	Topic string
	Data  []byte
}

// Error is the error type returned by a Producer when an error occurs while
// sending messages.
type Error struct {
	Msgs []*Message
	Err  error
}

// Producer represents a class that can send messages.
type Producer interface {
	// Name is the name of the producer.
	Name() string
	// Input is the message input channel.
	Input() chan<- *Message
	// Errors is the error output channel.
	Errors() <-chan *Error
	// Close closes the producer.
	Close() error
	// IsHealthy checks the health of the producer.
	IsHealthy() bool
}

// Consumer represents a class that can consume messages.
type Consumer interface {
	// Output gets messages until the given date.
	Output(endTime time.Time) (<-chan Messages, <-chan error)
	// Close closes the producer.
	Close() error
	// IsHealthy checks the health of the Consumer.
	IsHealthy() bool
}
