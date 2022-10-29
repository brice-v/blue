// Using this file to store the 'Global Object Store' basically things
// in go that cannot be easily translated to blue
package evaluator

import (
	"blue/object"
	"database/sql"
	"sync/atomic"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

var pidCount = atomic.Int64{}
var ProcessMap = &ConcurrentMap[int64, *object.Process]{
	kv: make(map[int64]*object.Process),
}

var dbCount = atomic.Int64{}
var DBMap = &ConcurrentMap[int64, *sql.DB]{
	kv: make(map[int64]*sql.DB),
}

var serverCount = atomic.Int64{}
var ServerMap = &ConcurrentMap[int64, *fiber.App]{
	kv: make(map[int64]*fiber.App),
}

var wsConnCount = atomic.Int64{}
var WSConnMap = &ConcurrentMap[int64, *websocket.Conn]{
	kv: make(map[int64]*websocket.Conn),
}
