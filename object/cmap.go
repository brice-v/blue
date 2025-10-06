package object

import "sync"

type ConcurrentMap[K comparable, V any] struct {
	Kv   map[K]V
	lock sync.RWMutex
}

func (cm *ConcurrentMap[K, V]) Put(k K, v V) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	cm.Kv[k] = v
}

func (cm *ConcurrentMap[K, V]) Get(k K) (V, bool) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	value, ok := cm.Kv[k]
	return value, ok
}

func (cm *ConcurrentMap[K, V]) GetNoCheck(k K) V {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	value := cm.Kv[k]
	return value
}

func (cm *ConcurrentMap[K, V]) GetAll() map[K]V {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	return cm.Kv
}

func (cm *ConcurrentMap[K, V]) Remove(k K) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	delete(cm.Kv, k)
}
