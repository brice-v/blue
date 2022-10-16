package evaluator

import (
	"blue/object"
	"database/sql"
	"sync"

	"github.com/gofiber/fiber/v2"
)

type ConcurrentMap[K comparable, V any] struct {
	kv   map[K]V
	lock sync.RWMutex
}

func NewPidMap() *ConcurrentMap[int64, *object.Process] {
	return &ConcurrentMap[int64, *object.Process]{
		kv: make(map[int64]*object.Process),
	}
}

func NewDBMap() *ConcurrentMap[int64, *sql.DB] {
	return &ConcurrentMap[int64, *sql.DB]{
		kv: make(map[int64]*sql.DB),
	}
}

func NewServerMap() *ConcurrentMap[int64, *fiber.App] {
	return &ConcurrentMap[int64, *fiber.App]{
		kv: make(map[int64]*fiber.App),
	}
}

type BuiltinMapType struct {
	*ConcurrentMap[string, *object.Builtin]
}

func NewBuiltinObjMap(input map[string]*object.Builtin) BuiltinMapType {
	return BuiltinMapType{&ConcurrentMap[string, *object.Builtin]{
		kv: input,
	}}
}

func (cm *ConcurrentMap[K, V]) Put(k K, v V) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	cm.kv[k] = v
}

func (cm *ConcurrentMap[K, V]) Get(k K) (V, bool) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	value, ok := cm.kv[k]
	return value, ok
}

func (cm *ConcurrentMap[K, V]) Remove(k K) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	delete(cm.kv, k)
}
