// Using this file to store the 'Global Object Store' basically things
// in go that cannot be easily translated to blue
package evaluator

import (
	"blue/evaluator/pubsub"
	"blue/object"
	"sync/atomic"

	"github.com/puzpuzpuz/xsync/v3"
)

var pidCount = atomic.Uint64{}

type ProcessKey struct {
	NodeName string
	Id       uint64
}

func (pk ProcessKey) Less(other ProcessKey) bool {
	return pk.Id < other.Id && pk.NodeName < other.NodeName
}

func (pk ProcessKey) Equal(other ProcessKey) bool {
	return pk.Id == other.Id && pk.NodeName == other.NodeName
}

func pk(name string, id uint64) ProcessKey {
	return ProcessKey{name, id}
}

var ProcessMap = xsync.NewMapOf[ProcessKey, *object.Process]()

var subscriberCount = atomic.Uint64{}
var PubSubBroker = pubsub.NewBroker()

var KVMap = xsync.NewMapOf[string, *object.Map]()

var GoObjId = atomic.Uint64{}
