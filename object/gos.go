package object

import (
	"embed"
	"sync/atomic"

	"github.com/puzpuzpuz/xsync/v3"
)

var PidCount = atomic.Uint64{}

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

func Pk(name string, id uint64) ProcessKey {
	return ProcessKey{name, id}
}

var ProcessMap = xsync.NewMapOf[ProcessKey, *Process]()

var SubscriberCount = atomic.Uint64{}
var PubSubBroker = NewBroker()

var KVMap = xsync.NewMapOf[string, *Map]()

var GoObjId = atomic.Uint64{}

var IsEmbed = false
var Files embed.FS

// NoExec is a global to prevent execution of shell commands on the system
var NoExec = false

func ClearGlobalState() {
	PidCount.Store(0)
	ProcessMap.Clear()
	SubscriberCount.Store(0)
	PubSubBroker.Clear()
	KVMap.Clear()
	GoObjId.Store(0)
}
