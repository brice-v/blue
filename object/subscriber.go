package object

import (
	"sync"
)

type Subscriber struct {
	id       uint64              // id of subscriber
	messages chan *Message       // messages channel
	topics   map[string]struct{} // topics it is subscribed to.
	active   bool                // if given subscriber is active
	mutex    sync.RWMutex        // lock
}

func createNewSubscriber(pid uint64) (uint64, *Subscriber) {
	return pid, &Subscriber{
		id:       pid,
		messages: make(chan *Message),
		topics:   map[string]struct{}{},
		active:   true,
	}
}

func (s *Subscriber) AddTopic(topic string) {
	// add topic to the subscriber
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	s.topics[topic] = struct{}{}
}

func (s *Subscriber) RemoveTopic(topic string) {
	// remove topic to the subscriber
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	delete(s.topics, topic)
}

func (s *Subscriber) destruct() {
	// destructor for subscriber.
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	s.active = false
	close(s.messages)
}

func (s *Subscriber) Signal(msg *Message) {
	// Gets the message from the channel
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if s.active {
		s.messages <- msg
	}
}

func (s *Subscriber) PollMessage() Object {
	// Listens to the message channel, prints once received.
	for {
		if msg, ok := <-s.messages; ok {
			mapObj := NewOrderedMap[string, Object]()
			mapObj.Set("topic", &Stringo{Value: msg.getTopic()})
			mapObj.Set("msg", msg.getMessage())
			return CreateMapObjectForGoMap(*mapObj)
		}
	}
}
