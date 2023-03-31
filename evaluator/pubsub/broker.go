package pubsub

import (
	"blue/object"
	"sync"
)

type Subscribers map[uint64]*Subscriber

type Broker struct {
	subscribers Subscribers            // map of subscribers id:Subscriber
	topics      map[string]Subscribers // map of topic to subscribers
	mut         sync.RWMutex           // mutex lock
}

func NewBroker() *Broker {
	// returns new broker object
	return &Broker{
		subscribers: Subscribers{},
		topics:      map[string]Subscribers{},
	}
}

func (b *Broker) GetTotalSubscribers() int {
	b.mut.RLock()
	defer b.mut.RUnlock()
	return len(b.subscribers)
}

func (b *Broker) GetNumSubscribersForTopic(topic string) int {
	b.mut.RLock()
	defer b.mut.RUnlock()
	return len(b.topics[topic])
}

func (b *Broker) AddSubscriber(pid uint64) *Subscriber {
	// Add subscriber to the broker.
	b.mut.Lock()
	defer b.mut.Unlock()
	id, s := createNewSubscriber(pid)
	b.subscribers[id] = s
	return s
}

func (b *Broker) RemoveSubscriber(s *Subscriber) {
	// remove subscriber to the broker.
	// unsubscribe to all topics which s is subscribed to.
	for topic := range s.topics {
		b.Unsubscribe(s, topic)
	}
	b.mut.Lock()
	// remove subscriber from list of subscribers.
	delete(b.subscribers, s.id)
	b.mut.Unlock()
	s.destruct()
}

func (b *Broker) Broadcast(msg object.Object, topics []string) {
	// broadcast message to all topics mentioned
	for _, topic := range topics {
		for _, s := range b.topics[topic] {
			m := NewMessage(topic, msg)
			go (func(s *Subscriber) {
				s.Signal(m)
			})(s)
		}
	}
}

func (b *Broker) BroadcastToAllTopics(msg object.Object) {
	// broadcast message to all topics mentioned
	for topic := range b.topics {
		for _, s := range b.topics[topic] {
			m := NewMessage(topic, msg)
			go (func(s *Subscriber) {
				s.Signal(m)
			})(s)
		}
	}
}

func (b *Broker) Subscribe(s *Subscriber, topic string) {
	// subscribe to given topic
	b.mut.Lock()
	defer b.mut.Unlock()

	if b.topics[topic] == nil {
		b.topics[topic] = Subscribers{}
	}
	s.AddTopic(topic)
	b.topics[topic][s.id] = s
}

func (b *Broker) Unsubscribe(s *Subscriber, topic string) {
	// unsubscribe to given topic
	b.mut.RLock()
	defer b.mut.RUnlock()

	delete(b.topics[topic], s.id)
	s.RemoveTopic(topic)
}

func (b *Broker) Publish(topic string, msg object.Object) {
	// publish the message to given topic.
	b.mut.RLock()
	bTopics := b.topics[topic]
	b.mut.RUnlock()
	for _, s := range bTopics {
		m := NewMessage(topic, msg)
		if !s.active {
			return
		}
		go (func(s *Subscriber) {
			s.Signal(m)
		})(s)
	}
}
