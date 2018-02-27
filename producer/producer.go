package producer

type Messages []*Message

type Message struct {
	Topic string
	Data  []byte
}

type Error struct {
	Msgs []*Message
	Err  error
}

type Producer interface {
	Name() string
	Input() chan<- *Message
	Errors() <-chan *Error
	Close() error
	IsHealthy() bool
}
