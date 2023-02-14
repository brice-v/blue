package object

import "sync"

type ConcurrentMap[K comparable, V any] struct {
	kv   map[K]V
	lock sync.RWMutex
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

func (cm *ConcurrentMap[K, V]) GetAll() map[K]V {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	return cm.kv
}

func (cm *ConcurrentMap[K, V]) Remove(k K) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	delete(cm.kv, k)
}
