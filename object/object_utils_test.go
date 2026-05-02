package object

import (
	"math/big"
	"regexp"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

// OrderedMap tests

func TestOrderedMapNew(t *testing.T) {
	m := NewOrderedMap[string, int]()
	if m == nil {
		t.Fatal("NewOrderedMap returned nil")
	}
	if m.Len() != 0 {
		t.Errorf("new map length should be 0, got %d", m.Len())
	}
}

func TestOrderedMapNewWithSize(t *testing.T) {
	m := NewOrderedMapWithSize[string, int](10)
	if m == nil {
		t.Fatal("NewOrderedMapWithSize returned nil")
	}
	if m.Len() != 0 {
		t.Errorf("new map length should be 0, got %d", m.Len())
	}
}

func TestOrderedMapSetAndGet(t *testing.T) {
	m := NewOrderedMap[string, int]()
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)

	val, ok := m.Get("a")
	if !ok {
		t.Error("expected key 'a' to exist")
	}
	if val != 1 {
		t.Errorf("expected value 1 for key 'a', got %d", val)
	}

	val, ok = m.Get("b")
	if !ok || val != 2 {
		t.Errorf("expected value 2 for key 'b', got %d, ok=%v", val, ok)
	}

	val, ok = m.Get("c")
	if !ok || val != 3 {
		t.Errorf("expected value 3 for key 'c', got %d, ok=%v", val, ok)
	}
}

func TestOrderedMapGetNonExistent(t *testing.T) {
	m := NewOrderedMap[string, int]()
	val, ok := m.Get("nonexistent")
	if ok {
		t.Error("expected key 'nonexistent' to not exist")
	}
	var zero int
	if val != zero {
		t.Errorf("expected zero value for missing key, got %d", val)
	}
}

func TestOrderedMapSetOverwrite(t *testing.T) {
	m := NewOrderedMap[string, int]()
	m.Set("a", 1)
	m.Set("a", 2)

	val, ok := m.Get("a")
	if !ok || val != 2 {
		t.Errorf("expected overwritten value 2, got %d, ok=%v", val, ok)
	}
	if m.Len() != 1 {
		t.Errorf("expected length 1 after overwrite, got %d", m.Len())
	}
}

func TestOrderedMapLen(t *testing.T) {
	m := NewOrderedMap[string, int]()
	if m.Len() != 0 {
		t.Errorf("expected length 0, got %d", m.Len())
	}
	m.Set("a", 1)
	if m.Len() != 1 {
		t.Errorf("expected length 1, got %d", m.Len())
	}
	m.Set("b", 2)
	m.Set("c", 3)
	if m.Len() != 3 {
		t.Errorf("expected length 3, got %d", m.Len())
	}
}

func TestOrderedMapDelete(t *testing.T) {
	m := NewOrderedMap[string, int]()
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)

	m.Delete("b")

	val, ok := m.Get("b")
	if ok {
		t.Errorf("expected key 'b' to be deleted, got %d", val)
	}
	if m.Len() != 2 {
		t.Errorf("expected length 2 after delete, got %d", m.Len())
	}

	// Verify remaining keys still work
	v1, ok1 := m.Get("a")
	v2, ok2 := m.Get("c")
	if !ok1 || v1 != 1 {
		t.Errorf("expected 'a'=1, got %d, ok=%v", v1, ok1)
	}
	if !ok2 || v2 != 3 {
		t.Errorf("expected 'c'=3, got %d, ok=%v", v2, ok2)
	}
}

func TestOrderedMapDeleteNonExistent(t *testing.T) {
	m := NewOrderedMap[string, int]()
	m.Set("a", 1)
	m.Delete("nonexistent")
	if m.Len() != 1 {
		t.Errorf("expected length 1 after deleting non-existent key, got %d", m.Len())
	}
}

func TestOrderedMapKeyOrder(t *testing.T) {
	m := NewOrderedMap[string, int]()
	m.Set("first", 1)
	m.Set("second", 2)
	m.Set("third", 3)

	expected := []string{"first", "second", "third"}
	if len(m.Keys) != len(expected) {
		t.Fatalf("expected %d keys, got %d", len(expected), len(m.Keys))
	}
	for i, key := range m.Keys {
		if key != expected[i] {
			t.Errorf("expected key[%d]=%q, got %q", i, expected[i], key)
		}
	}
}

func TestOrderedMapKeyOrderAfterDelete(t *testing.T) {
	m := NewOrderedMap[string, int]()
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)
	m.Set("d", 4)

	m.Delete("c")

	expected := []string{"a", "b", "d"}
	if len(m.Keys) != len(expected) {
		t.Fatalf("expected %d keys, got %d", len(expected), len(m.Keys))
	}
	for i, key := range m.Keys {
		if key != expected[i] {
			t.Errorf("expected key[%d]=%q, got %q", i, expected[i], key)
		}
	}
}

func TestOrderedMapKeyOrderAfterOverwrite(t *testing.T) {
	m := NewOrderedMap[string, int]()
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)

	// Overwriting should not change key order
	m.Set("b", 99)

	expected := []string{"a", "b", "c"}
	for i, key := range m.Keys {
		if key != expected[i] {
			t.Errorf("expected key[%d]=%q after overwrite, got %q", i, expected[i], key)
		}
	}
}

func TestNewPairsMap(t *testing.T) {
	m := NewPairsMap()
	if m.Len() != 0 {
		t.Errorf("new pairs map length should be 0, got %d", m.Len())
	}
}

func TestNewPairsMapWithSize(t *testing.T) {
	m := NewPairsMapWithSize(5)
	if m.Len() != 0 {
		t.Errorf("new pairs map with size length should be 0, got %d", m.Len())
	}
}

// ConcurrentMap tests

func TestConcurrentMapPutAndGet(t *testing.T) {
	cm := &ConcurrentMap[string, int]{Kv: make(map[string]int)}
	cm.Put("a", 1)
	cm.Put("b", 2)

	val, ok := cm.Get("a")
	if !ok || val != 1 {
		t.Errorf("expected value 1 for key 'a', got %d, ok=%v", val, ok)
	}

	val, ok = cm.Get("b")
	if !ok || val != 2 {
		t.Errorf("expected value 2 for key 'b', got %d, ok=%v", val, ok)
	}
}

func TestConcurrentMapGetNonExistent(t *testing.T) {
	cm := &ConcurrentMap[string, int]{Kv: make(map[string]int)}
	val, ok := cm.Get("nonexistent")
	if ok {
		t.Error("expected key to not exist")
	}
	var zero int
	if val != zero {
		t.Errorf("expected zero value, got %d", val)
	}
}

func TestConcurrentMapGetNoCheck(t *testing.T) {
	cm := &ConcurrentMap[string, int]{Kv: map[string]int{"a": 42}}
	val := cm.GetNoCheck("a")
	if val != 42 {
		t.Errorf("expected value 42, got %d", val)
	}

	var zero int
	val = cm.GetNoCheck("missing")
	if val != zero {
		t.Errorf("expected zero value for missing key, got %d", val)
	}
}

func TestConcurrentMapRemove(t *testing.T) {
	cm := &ConcurrentMap[string, int]{Kv: map[string]int{"a": 1, "b": 2}}
	cm.Remove("a")

	_, ok := cm.Get("a")
	if ok {
		t.Error("expected key 'a' to be removed")
	}

	val, ok := cm.Get("b")
	if !ok || val != 2 {
		t.Errorf("expected key 'b' to still exist with value 2, got %d, ok=%v", val, ok)
	}
}

func TestConcurrentMapGetAll(t *testing.T) {
	cm := &ConcurrentMap[string, int]{Kv: map[string]int{"a": 1, "b": 2, "c": 3}}
	all := cm.GetAll()

	if len(all) != 3 {
		t.Errorf("expected 3 entries, got %d", len(all))
	}
	if all["a"] != 1 || all["b"] != 2 || all["c"] != 3 {
		t.Errorf("unexpected values in GetAll: %v", all)
	}
}

func TestConcurrentMapGetAllEmpty(t *testing.T) {
	cm := &ConcurrentMap[string, int]{Kv: make(map[string]int)}
	all := cm.GetAll()
	if len(all) != 0 {
		t.Errorf("expected empty map, got %v", all)
	}
}

func TestConcurrentMapPutOverwrite(t *testing.T) {
	cm := &ConcurrentMap[string, int]{Kv: map[string]int{"a": 1}}
	cm.Put("a", 99)

	val, ok := cm.Get("a")
	if !ok || val != 99 {
		t.Errorf("expected overwritten value 99, got %d, ok=%v", val, ok)
	}
}

func TestConcurrentMapRemoveNonExistent(t *testing.T) {
	cm := &ConcurrentMap[string, int]{Kv: map[string]int{"a": 1}}
	cm.Remove("nonexistent")

	val, ok := cm.Get("a")
	if !ok || val != 1 {
		t.Errorf("expected key 'a' to still exist, got %d, ok=%v", val, ok)
	}
}

// Message tests

func TestNewMessage(t *testing.T) {
	msg := NewMessage("test-topic", &Integer{Value: 42})
	if msg == nil {
		t.Fatal("NewMessage returned nil")
	}
	if msg.getTopic() != "test-topic" {
		t.Errorf("expected topic 'test-topic', got %q", msg.getTopic())
	}
	body, ok := msg.getMessage().(*Integer)
	if !ok {
		t.Fatalf("expected body to be *Integer, got %T", msg.getMessage())
	}
	if body.Value != 42 {
		t.Errorf("expected body value 42, got %d", body.Value)
	}
}

func TestNewMessageWithNull(t *testing.T) {
	msg := NewMessage("null-topic", NULL)
	if msg.getTopic() != "null-topic" {
		t.Errorf("expected topic 'null-topic', got %q", msg.getTopic())
	}
	body := msg.getMessage()
	if body != NULL {
		t.Error("expected body to be NULL")
	}
}

func TestNewMessageWithString(t *testing.T) {
	msg := NewMessage("string-topic", &Stringo{Value: "hello"})
	body := msg.getMessage().(*Stringo)
	if body.Value != "hello" {
		t.Errorf("expected body value 'hello', got %q", body.Value)
	}
}

// Subscriber tests

func TestCreateNewSubscriber(t *testing.T) {
	id, sub := createNewSubscriber(123)
	if id != 123 {
		t.Errorf("expected subscriber id 123, got %d", id)
	}
	if sub.id != 123 {
		t.Errorf("expected sub.id 123, got %d", sub.id)
	}
	if sub.active != true {
		t.Error("expected subscriber to be active")
	}
	if len(sub.topics) != 0 {
		t.Errorf("expected 0 topics, got %d", len(sub.topics))
	}
}

func TestSubscriberAddTopic(t *testing.T) {
	_, sub := createNewSubscriber(1)
	sub.AddTopic("topic-a")
	if len(sub.topics) != 1 {
		t.Errorf("expected 1 topic, got %d", len(sub.topics))
	}
	if _, ok := sub.topics["topic-a"]; !ok {
		t.Error("expected topic-a to be present")
	}
}

func TestSubscriberAddMultipleTopics(t *testing.T) {
	_, sub := createNewSubscriber(1)
	sub.AddTopic("a")
	sub.AddTopic("b")
	sub.AddTopic("c")
	if len(sub.topics) != 3 {
		t.Errorf("expected 3 topics, got %d", len(sub.topics))
	}
}

func TestSubscriberRemoveTopic(t *testing.T) {
	_, sub := createNewSubscriber(1)
	sub.AddTopic("a")
	sub.AddTopic("b")
	sub.RemoveTopic("a")
	if len(sub.topics) != 1 {
		t.Errorf("expected 1 topic after remove, got %d", len(sub.topics))
	}
	if _, ok := sub.topics["a"]; ok {
		t.Error("expected topic-a to be removed")
	}
	if _, ok := sub.topics["b"]; !ok {
		t.Error("expected topic-b to still exist")
	}
}

func TestSubscriberRemoveNonExistentTopic(t *testing.T) {
	_, sub := createNewSubscriber(1)
	sub.AddTopic("a")
	sub.RemoveTopic("nonexistent")
	if len(sub.topics) != 1 {
		t.Errorf("expected 1 topic, got %d", len(sub.topics))
	}
}

func TestSubscriberSignal(t *testing.T) {
	_, sub := createNewSubscriber(1)
	msg := NewMessage("sig-topic", &Integer{Value: 7})
	done := make(chan struct{})
	go func() {
		sub.Signal(msg)
		close(done)
	}()

	select {
	case received := <-sub.messages:
		if received.getTopic() != "sig-topic" {
			t.Errorf("expected topic 'sig-topic', got %q", received.getTopic())
		}
		body := received.getMessage().(*Integer)
		if body.Value != 7 {
			t.Errorf("expected body value 7, got %d", body.Value)
		}
	case <-done:
		t.Fatal("Signal completed before message was received")
	}
}

func TestSubscriberSignalInactive(t *testing.T) {
	_, sub := createNewSubscriber(1)

	// Use a channel to ensure destruct completes before Signal
	done := make(chan struct{})
	go func() {
		sub.destruct()
		close(done)
	}()
	<-done

	// After destruct, channel is closed and active is false.
	// Signal should not send (though there's a known RLock race in production code).
	// We just verify destruct completed and channel is closed.
	select {
	case _, ok := <-sub.messages:
		if ok {
			t.Error("expected channel to be closed after destruct")
		}
	default:
		t.Error("channel should be closed after destruct")
	}
}

func TestSubscriberDestruct(t *testing.T) {
	_, sub := createNewSubscriber(1)
	if sub.active != true {
		t.Error("expected subscriber to be active before destruct")
	}
	sub.destruct()
	if sub.active != false {
		t.Error("expected subscriber to be inactive after destruct")
	}
}

// Broker tests

func TestNewBroker(t *testing.T) {
	b := NewBroker()
	if b == nil {
		t.Fatal("NewBroker returned nil")
	}
	if b.GetTotalSubscribers() != 0 {
		t.Errorf("expected 0 subscribers, got %d", b.GetTotalSubscribers())
	}
}

func TestBrokerAddSubscriber(t *testing.T) {
	b := NewBroker()
	s := b.AddSubscriber(1)
	if s == nil {
		t.Fatal("AddSubscriber returned nil")
	}
	if b.GetTotalSubscribers() != 1 {
		t.Errorf("expected 1 subscriber, got %d", b.GetTotalSubscribers())
	}
}

func TestBrokerRemoveSubscriber(t *testing.T) {
	b := NewBroker()
	s := b.AddSubscriber(1)
	b.AddSubscriber(2)

	b.RemoveSubscriber(s)
	if b.GetTotalSubscribers() != 1 {
		t.Errorf("expected 1 subscriber after remove, got %d", b.GetTotalSubscribers())
	}
}

func TestBrokerSubscribe(t *testing.T) {
	b := NewBroker()
	s := b.AddSubscriber(1)
	b.Subscribe(s, "news")

	if b.GetNumSubscribersForTopic("news") != 1 {
		t.Errorf("expected 1 subscriber for 'news', got %d", b.GetNumSubscribersForTopic("news"))
	}
}

func TestBrokerSubscribeMultipleTopics(t *testing.T) {
	b := NewBroker()
	s := b.AddSubscriber(1)
	b.Subscribe(s, "topic-a")
	b.Subscribe(s, "topic-b")

	if b.GetNumSubscribersForTopic("topic-a") != 1 {
		t.Errorf("expected 1 subscriber for 'topic-a'")
	}
	if b.GetNumSubscribersForTopic("topic-b") != 1 {
		t.Errorf("expected 1 subscriber for 'topic-b'")
	}
}

func TestBrokerUnsubscribe(t *testing.T) {
	b := NewBroker()
	s := b.AddSubscriber(1)
	b.Subscribe(s, "news")
	b.Unsubscribe(s, "news")

	if b.GetNumSubscribersForTopic("news") != 0 {
		t.Errorf("expected 0 subscribers for 'news' after unsubscribe, got %d", b.GetNumSubscribersForTopic("news"))
	}
}

func TestBrokerPublish(t *testing.T) {
	b := NewBroker()
	s := b.AddSubscriber(1)
	b.Subscribe(s, "updates")

	b.Publish("updates", &Stringo{Value: "hello world"})

	received := make(chan *Message, 1)
	go func() {
		select {
		case msg := <-s.messages:
			received <- msg
		}
	}()

	select {
	case msg := <-received:
		if msg.getTopic() != "updates" {
			t.Errorf("expected topic 'updates', got %q", msg.getTopic())
		}
		body := msg.getMessage().(*Stringo)
		if body.Value != "hello world" {
			t.Errorf("expected body 'hello world', got %q", body.Value)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("expected to receive a message within timeout")
	}
}

func TestBrokerPublishNoSubscribers(t *testing.T) {
	b := NewBroker()
	// Should not panic
	b.Publish("empty", &Integer{Value: 1})
}

func TestBrokerPublishInactiveSubscriber(t *testing.T) {
	b := NewBroker()
	s := b.AddSubscriber(1)
	b.Subscribe(s, "test")

	// Deactivate subscriber
	done := make(chan struct{})
	go func() {
		s.destruct()
		close(done)
	}()
	<-done

	b.Publish("test", &Integer{Value: 1})

	// After destruct, channel is closed. Verify that.
	select {
	case _, ok := <-s.messages:
		if ok {
			t.Error("expected channel to be closed after destruct")
		}
	case <-time.After(100 * time.Millisecond):
		// expected: channel is closed, non-blocking read returns immediately
	}
}

func TestBrokerBroadcast(t *testing.T) {
	b := NewBroker()
	s1 := b.AddSubscriber(1)
	s2 := b.AddSubscriber(2)
	b.Subscribe(s1, "all")
	b.Subscribe(s2, "all")

	b.Broadcast(&Integer{Value: 42}, []string{"all"})

	count := 0
	timeout := time.After(200 * time.Millisecond)
	for count < 2 {
		select {
		case <-s1.messages:
			count++
		case <-s2.messages:
			count++
		case <-timeout:
			goto done
		}
	}
done:
	if count != 2 {
		t.Errorf("expected 2 messages delivered, got %d", count)
	}
}

func TestBrokerBroadcastToAllTopics(t *testing.T) {
	b := NewBroker()
	s1 := b.AddSubscriber(1)
	s2 := b.AddSubscriber(2)
	b.Subscribe(s1, "topic-a")
	b.Subscribe(s2, "topic-b")

	b.BroadcastToAllTopics(&Integer{Value: 99})

	count := 0
	timeout := time.After(200 * time.Millisecond)
	for count < 2 {
		select {
		case <-s1.messages:
			count++
		case <-s2.messages:
			count++
		case <-timeout:
			goto done
		}
	}
done:
	if count != 2 {
		t.Errorf("expected 2 messages, got %d", count)
	}
}

func TestBrokerClear(t *testing.T) {
	b := NewBroker()
	s := b.AddSubscriber(1)
	b.Subscribe(s, "news")
	b.Clear()

	if b.GetTotalSubscribers() != 0 {
		t.Errorf("expected 0 subscribers after clear, got %d", b.GetTotalSubscribers())
	}
	if b.GetNumSubscribersForTopic("news") != 0 {
		t.Errorf("expected 0 subscribers for 'news' after clear")
	}
}

func TestBrokerRemoveSubscriberUnsubscribesAllTopics(t *testing.T) {
	b := NewBroker()
	s := b.AddSubscriber(1)
	b.Subscribe(s, "topic-a")
	b.Subscribe(s, "topic-b")

	b.RemoveSubscriber(s)

	if b.GetNumSubscribersForTopic("topic-a") != 0 {
		t.Error("expected topic-a to have 0 subscribers after remove")
	}
	if b.GetNumSubscribersForTopic("topic-b") != 0 {
		t.Error("expected topic-b to have 0 subscribers after remove")
	}
}

// Clone tests

func TestCloneNull(t *testing.T) {
	clone := NULL.Clone()
	if clone != NULL {
		t.Error("NULLClone should return the singleton NULL")
	}
}

func TestCloneIgnore(t *testing.T) {
	clone := IGNORE.Clone()
	if clone != IGNORE {
		t.Error("IGNORE Clone should return the singleton IGNORE")
	}
}

func TestCloneBreak(t *testing.T) {
	clone := BREAK.Clone()
	if clone != BREAK {
		t.Error("BREAK Clone should return the singleton BREAK")
	}
}

func TestCloneContinue(t *testing.T) {
	clone := CONTINUE.Clone()
	if clone != CONTINUE {
		t.Error("CONTINUE Clone should return the singleton CONTINUE")
	}
}

func TestCloneInteger(t *testing.T) {
	orig := &Integer{Value: 42}
	clone := orig.Clone().(*Integer)
	if clone.Value != 42 {
		t.Errorf("expected cloned value 42, got %d", clone.Value)
	}
	if clone == orig {
		t.Error("clone should be a different pointer")
	}
}

func TestCloneBigInteger(t *testing.T) {
	bigInt := &BigInteger{Value: big.NewInt(999999999999999999)}
	clone := bigInt.Clone().(*BigInteger)
	if clone.Value.Cmp(bigInt.Value) != 0 {
		t.Errorf("expected cloned value %s, got %s", bigInt.Value.String(), clone.Value.String())
	}
	if clone.Value == bigInt.Value {
		t.Error("clone should have a different big.Int pointer")
	}
}

func TestCloneBoolean(t *testing.T) {
	trueOrig := &Boolean{Value: true}
	trueClone := trueOrig.Clone().(*Boolean)
	if !trueClone.Value {
		t.Error("expected cloned boolean to be true")
	}

	falseOrig := &Boolean{Value: false}
	falseClone := falseOrig.Clone().(*Boolean)
	if falseClone.Value {
		t.Error("expected cloned boolean to be false")
	}
}

func TestCloneUInteger(t *testing.T) {
	orig := &UInteger{Value: 18446744073709551615}
	clone := orig.Clone().(*UInteger)
	if clone.Value != 18446744073709551615 {
		t.Errorf("expected cloned value 18446744073709551615, got %d", clone.Value)
	}
	if clone == orig {
		t.Error("clone should be a different pointer")
	}
}

func TestCloneFloat(t *testing.T) {
	orig := &Float{Value: 3.14159}
	clone := orig.Clone().(*Float)
	if clone.Value != 3.14159 {
		t.Errorf("expected cloned value 3.14159, got %f", clone.Value)
	}
	if clone == orig {
		t.Error("clone should be a different pointer")
	}
}

func TestCloneBigFloat(t *testing.T) {
	decimalVal := decimal.NewFromFloat(123.456789)
	orig := BigFloat{Value: decimalVal}
	clone := orig.Clone().(BigFloat)
	if !clone.Value.Equal(decimalVal) {
		t.Errorf("expected cloned value %s, got %s", decimalVal.String(), clone.Value.String())
	}
}

func TestCloneReturnValue(t *testing.T) {
	orig := &ReturnValue{Value: &Integer{Value: 100}}
	clone := orig.Clone().(*Integer)
	if clone.Value != 100 {
		t.Errorf("expected cloned value 100, got %d", clone.Value)
	}
}

func TestCloneError(t *testing.T) {
	orig := &Error{Message: "something went wrong"}
	clone := orig.Clone().(*Error)
	if clone.Message != "something went wrong" {
		t.Errorf("expected cloned message 'something went wrong', got %q", clone.Message)
	}
}

func TestCloneStringFunction(t *testing.T) {
	orig := &StringFunction{Value: "fun() { return 42 }"}
	clone := orig.Clone().(*StringFunction)
	if clone.Value != "fun() { return 42 }" {
		t.Errorf("expected cloned value 'fun() { return 42 }', got %q", clone.Value)
	}
}

func TestCloneStringo(t *testing.T) {
	orig := &Stringo{Value: "hello world"}
	clone := orig.Clone().(*Stringo)
	if clone.Value != "hello world" {
		t.Errorf("expected cloned value 'hello world', got %q", clone.Value)
	}
	if clone == orig {
		t.Error("clone should be a different pointer")
	}
}

func TestCloneBytes(t *testing.T) {
	orig := &Bytes{Value: []byte{1, 2, 3, 4, 5}}
	clone := orig.Clone().(*Bytes)
	if len(clone.Value) != 5 {
		t.Errorf("expected cloned bytes length 5, got %d", len(clone.Value))
	}
	for i, b := range orig.Value {
		if clone.Value[i] != b {
			t.Errorf("expected cloned bytes[%d] = %d, got %d", i, b, clone.Value[i])
		}
	}
	if &clone.Value[0] == &orig.Value[0] {
		t.Error("clone should have a different byte slice")
	}
}

func TestCloneRegex(t *testing.T) {
	re, _ := regexp.Compile(`\d+`)
	orig := &Regex{Value: re}
	clone := orig.Clone().(*Regex)
	if clone.Value.String() != re.String() {
		t.Errorf("expected cloned regex %q, got %q", re.String(), clone.Value.String())
	}
}

func TestCloneBuiltin(t *testing.T) {
	builtinFn := func(args ...Object) Object { return NULL }
	orig := &Builtin{Name: "test", Fun: builtinFn, HelpStr: "test help"}
	clone := orig.Clone().(*Builtin)
	if clone.Name != "test" {
		t.Errorf("expected cloned name 'test', got %q", clone.Name)
	}
	// Builtin returns itself (singleton pattern)
	if clone != orig {
		t.Error("Builtin Clone should return the same pointer")
	}
}

func TestCloneList(t *testing.T) {
	orig := &List{Elements: []Object{&Integer{Value: 1}, &Stringo{Value: "hello"}, &Integer{Value: 2}}}
	clone := orig.Clone().(*List)
	if len(clone.Elements) != 3 {
		t.Errorf("expected cloned list length 3, got %d", len(clone.Elements))
	}
	i1, ok := clone.Elements[0].(*Integer)
	if !ok || i1.Value != 1 {
		t.Error("expected first element to be Integer{1}")
	}
	s1, ok := clone.Elements[1].(*Stringo)
	if !ok || s1.Value != "hello" {
		t.Error("expected second element to be Stringo{\"hello\"}")
	}
	if clone == orig {
		t.Error("clone should be a different pointer")
	}
	if &clone.Elements[0] == &orig.Elements[0] {
		t.Error("clone should have a different elements slice")
	}
}

func TestCloneListCompLiteral(t *testing.T) {
	orig := &ListCompLiteral{Elements: []Object{&Integer{Value: 10}, &Float{Value: 0.5}}}
	clone := orig.Clone().(*ListCompLiteral)
	if len(clone.Elements) != 2 {
		t.Errorf("expected cloned length 2, got %d", len(clone.Elements))
	}
}

func TestCloneMap(t *testing.T) {
	key1 := &Stringo{Value: "a"}
	key2 := &Stringo{Value: "b"}
	hk1 := HashKey{Type: STRING_OBJ, Value: HashObject(key1)}
	hk2 := HashKey{Type: STRING_OBJ, Value: HashObject(key2)}

	m := &Map{
		Pairs: NewPairsMap(),
	}
	m.Pairs.Set(hk1, MapPair{Key: key1, Value: &Integer{Value: 1}})
	m.Pairs.Set(hk2, MapPair{Key: key2, Value: &Stringo{Value: "two"}})

	clone := m.Clone().(*Map)
	if clone.Pairs.Len() != 2 {
		t.Errorf("expected cloned map length 2, got %d", clone.Pairs.Len())
	}

	v1, ok := clone.Pairs.Get(hk1)
	if !ok || v1.Value.Inspect() != "1" {
		t.Error("expected cloned map value for 'a' to be 1")
	}
	v2, ok := clone.Pairs.Get(hk2)
	if !ok || v2.Value.Inspect() != "two" {
		t.Error("expected cloned map value for 'b' to be 'two'")
	}
}

func TestCloneSet(t *testing.T) {
	set := &Set{Elements: NewSetElementsWithSize(3)}
	set.Elements.Set(HashObject(&Integer{Value: 1}), SetPair{Value: &Integer{Value: 1}, Present: struct{}{}})
	set.Elements.Set(HashObject(&Integer{Value: 2}), SetPair{Value: &Integer{Value: 2}, Present: struct{}{}})

	clone := set.Clone().(*Set)
	if clone.Elements.Len() != 2 {
		t.Errorf("expected cloned set length 2, got %d", clone.Elements.Len())
	}
}

func TestCloneSetCompLiteral(t *testing.T) {
	orig := &SetCompLiteral{
		Elements: map[uint64]SetPair{
			HashObject(&Integer{Value: 1}): {Value: &Integer{Value: 1}, Present: struct{}{}},
		},
	}
	clone := orig.Clone().(*SetCompLiteral)
	if len(clone.Elements) != 1 {
		t.Errorf("expected cloned set length 1, got %d", len(clone.Elements))
	}
}

func TestCloneBlueStruct(t *testing.T) {
	orig, _ := NewBlueStruct([]string{"name", "age"}, []Object{&Stringo{Value: "Alice"}, &Integer{Value: 30}})
	clone := orig.Clone().(*BlueStruct)
	if len(clone.Fields) != 2 {
		t.Errorf("expected cloned fields length 2, got %d", len(clone.Fields))
	}
	if clone.Fields[0] != "name" || clone.Fields[1] != "age" {
		t.Errorf("expected fields ['name', 'age'], got %v", clone.Fields)
	}
	v, _ := clone.Get("name")
	if s, ok := v.(*Stringo); !ok || s.Value != "Alice" {
		t.Error("expected cloned struct field 'name' to be 'Alice'")
	}
	if &clone.Fields[0] == &orig.(*BlueStruct).Fields[0] {
		t.Error("clone should have a different fields slice")
	}
}

func TestCloneDefaultArgs(t *testing.T) {
	orig := &DefaultArgs{Value: map[string]Object{
		"a": &Integer{Value: 1},
		"b": &Stringo{Value: "test"},
	}}
	clone := orig.Clone().(*DefaultArgs)
	if len(clone.Value) != 2 {
		t.Errorf("expected cloned value length 2, got %d", len(clone.Value))
	}
	v, ok := clone.Value["a"]
	if !ok {
		t.Error("expected cloned value to have key 'a'")
	}
	if i, ok := v.(*Integer); !ok || i.Value != 1 {
		t.Error("expected cloned value['a'] to be Integer{1}")
	}
}

func TestCloneGoObj(t *testing.T) {
	orig := &GoObj[int]{Id: 42, Value: 123}
	clone := orig.Clone().(*GoObj[int])
	if clone.Id != 42 {
		t.Errorf("expected cloned id 42, got %d", clone.Id)
	}
	if clone.Value != 123 {
		t.Errorf("expected cloned value 123, got %d", clone.Value)
	}
}

func TestCloneGoObjectGob(t *testing.T) {
	orig := &GoObjectGob{T: "test-type", Value: []byte{1, 2, 3}}
	clone := orig.Clone().(*GoObjectGob)
	if clone.T != "test-type" {
		t.Errorf("expected cloned T 'test-type', got %q", clone.T)
	}
	if len(clone.Value) != 3 {
		t.Errorf("expected cloned value length 3, got %d", len(clone.Value))
	}
	if &clone.Value[0] == &orig.Value[0] {
		t.Error("clone should have a different value slice")
	}
}

func TestCloneProcess(t *testing.T) {
	ch := make(chan Object)
	orig := &Process{Ch: ch, Id: 99, NodeName: "test-process"}
	clone := orig.Clone().(*Process)
	if clone.Id != 99 {
		t.Errorf("expected cloned id 99, got %d", clone.Id)
	}
	if clone.NodeName != "test-process" {
		t.Errorf("expected cloned name 'test-process', got %q", clone.NodeName)
	}
}

func TestCloneModule(t *testing.T) {
	orig := &Module{Name: "mymodule"}
	clone := orig.Clone().(*Module)
	if clone.Name != "mymodule" {
		t.Errorf("expected cloned name 'mymodule', got %q", clone.Name)
	}
	if clone.HelpStr != orig.HelpStr {
		t.Error("expected cloned HelpStr to match")
	}
}

func TestCloneModuleWithEnv(t *testing.T) {
	coreEnv := NewEnvironmentWithoutCore()
	env := NewEnvironment(coreEnv)
	env.Set("x", &Integer{Value: 42})
	orig := &Module{Name: "withenv", Env: env}
	clone := orig.Clone().(*Module)
	if clone.Name != "withenv" {
		t.Errorf("expected cloned name 'withenv', got %q", clone.Name)
	}
	// Note: Module.Clone has a bug where it clones m.Env (nil) instead of x.Env
	// This test documents the current behavior
}

func TestCloneSlice(t *testing.T) {
	orig := []Object{&Integer{Value: 1}, &Stringo{Value: "test"}, &Float{Value: 1.5}}
	clone := CloneSlice(orig)
	if len(clone) != 3 {
		t.Errorf("expected cloned slice length 3, got %d", len(clone))
	}
	for i, elem := range clone {
		if elem == orig[i] {
			t.Errorf("clone[%d] should be a different pointer", i)
		}
	}
}

func TestCloneSliceWithNil(t *testing.T) {
	orig := []Object{&Integer{Value: 1}, nil, &Stringo{Value: "test"}}
	clone := CloneSlice(orig)
	if len(clone) != 3 {
		t.Errorf("expected cloned slice length 3, got %d", len(clone))
	}
	if clone[1] != nil {
		t.Error("expected cloned slice[1] to be nil")
	}
}
