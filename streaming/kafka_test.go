package streaming

import (
	"testing"
	"github.com/magiconair/properties/assert"
	"github.com/Shopify/sarama"
)

func Test_newProducerMessage(t *testing.T) {
	m := &Message{
		Topic: "topic",
		Key:   []byte("key"),
		Data:  []byte("data"),
	}

	pm := newProducerMessage(m)

	assert.Equal(t, pm.Topic, "topic")
	assert.Equal(t, pm.Key, sarama.ByteEncoder("key"))
	assert.Equal(t, pm.Value, sarama.ByteEncoder("data"))
}

func Test_newProducerMessageWithEmptyKey(t *testing.T) {
	m := &Message{
		Topic: "topic",
		Key:   []byte(""),
		Data:  []byte("data"),
	}

	pm := newProducerMessage(m)

	assert.Equal(t, pm.Topic, "topic")
	assert.Equal(t, pm.Key, nil)
	assert.Equal(t, pm.Value, sarama.ByteEncoder("data"))
}
