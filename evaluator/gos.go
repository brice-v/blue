// Using this file to store the 'Global Object Store' basically things
// in go that cannot be easily translated to blue
package evaluator

import (
	"blue/evaluator/pubsub"
	"blue/object"
	"sync/atomic"
)

var pidCount = atomic.Uint64{}
var ProcessMap = &ConcurrentMap[uint64, *object.Process]{
	kv: make(map[uint64]*object.Process),
}

var subscriberCount = atomic.Uint64{}
var PubSubBroker = pubsub.NewBroker()

var KVMap = &ConcurrentMap[string, *object.Map]{
	kv: make(map[string]*object.Map),
}

var GoObjId = atomic.Uint64{}
