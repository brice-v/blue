package pubsub

import "blue/object"

type Message struct {
	topic string
	body  object.Object
}

func NewMessage(topic string, msg object.Object) *Message {
	// Returns the message object
	return &Message{
		topic: topic,
		body:  msg,
	}
}

func (m *Message) getTopic() string {
	// returns the topic of the message
	return m.topic
}

func (m *Message) getMessage() object.Object {
	// returns the message body.
	return m.body
}
