// Using this file to store the 'Global Object Store' basically things
// in go that cannot be easily translated to blue
package evaluator

import (
	"blue/evaluator/pubsub"
	"blue/object"
	"database/sql"
	"net"
	"sync/atomic"

	"fyne.io/fyne/v2"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/gookit/color"
	ws "github.com/gorilla/websocket"
)

var pidCount = atomic.Uint64{}
var ProcessMap = &ConcurrentMap[uint64, *object.Process]{
	kv: make(map[uint64]*object.Process),
}

var dbCount = atomic.Uint64{}
var DBMap = &ConcurrentMap[uint64, *sql.DB]{
	kv: make(map[uint64]*sql.DB),
}

var serverCount = atomic.Uint64{}
var ServerMap = &ConcurrentMap[uint64, *fiber.App]{
	kv: make(map[uint64]*fiber.App),
}

var wsConnCount = atomic.Uint64{}
var WSConnMap = &ConcurrentMap[uint64, *websocket.Conn]{
	kv: make(map[uint64]*websocket.Conn),
}

var wsClientConnCount = atomic.Uint64{}
var WSClientConnMap = &ConcurrentMap[uint64, *ws.Conn]{
	kv: make(map[uint64]*ws.Conn),
}

var netConnCount = atomic.Uint64{}
var NetConnMap = &ConcurrentMap[uint64, net.Conn]{
	kv: make(map[uint64]net.Conn),
}

var netTCPServerCount = atomic.Uint64{}
var NetTCPServerMap = &ConcurrentMap[uint64, net.Listener]{
	kv: make(map[uint64]net.Listener),
}

var netUDPServerCount = atomic.Uint64{}
var NetUDPServerMap = &ConcurrentMap[uint64, *net.UDPConn]{
	kv: make(map[uint64]*net.UDPConn),
}

// UI Object stores
var uiAppCount = atomic.Uint64{}
var UIAppMap = &ConcurrentMap[uint64, fyne.App]{
	kv: make(map[uint64]fyne.App),
}

var uiCanvasObjectCount = atomic.Uint64{}
var UICanvasObjectMap = &ConcurrentMap[uint64, fyne.CanvasObject]{
	kv: make(map[uint64]fyne.CanvasObject),
}

var colorStyleCount = atomic.Uint64{}
var ColorStyleMap = &ConcurrentMap[uint64, color.Style]{
	kv: make(map[uint64]color.Style),
}
var ColorStyleCountMap = &ConcurrentMap[string, uint64]{
	kv: make(map[string]uint64),
}

var SubscriberMap = &ConcurrentMap[uint64, *pubsub.Subscriber]{
	kv: make(map[uint64]*pubsub.Subscriber),
}
var PubSubBroker = pubsub.NewBroker()
