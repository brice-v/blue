package pubsub

import (
	"blue/object"
	"sort"
	"sync"
)

type Subscriber struct {
	id       uint64              // id of subscriber
	messages chan *Message       // messages channel
	topics   map[string]struct{} // topics it is subscribed to.
	active   bool                // if given subscriber is active
	mutex    sync.RWMutex        // lock
}

func CreateNewSubscriber(pid uint64) (uint64, *Subscriber) {
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

// These are all things

func (s *Subscriber) GetTopics() []string {
	// Get all topic of the subscriber
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	topics := make([]string, len(s.topics))
	i := 0
	for topic := range s.topics {
		topics[i] = topic
		i++
	}
	sort.Strings(topics)
	return topics
}

func (s *Subscriber) Destruct() {
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

func (s *Subscriber) PollMessage() object.Object {
	// Listens to the message channel, prints once received.
	for {
		if msg, ok := <-s.messages; ok {
			mapObj := object.NewOrderedMap[string, object.Object]()
			mapObj.Set("topic", &object.Stringo{Value: msg.GetTopic()})
			mapObj.Set("msg", msg.GetMessage())
			return object.CreateMapObjectForGoMap(*mapObj)
		}
	}
}
