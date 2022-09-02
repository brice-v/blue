package evaluator

import (
	"blue/object"
	"sync"
)

type ConcurrentMap struct {
	kv   map[int64]*object.Process
	lock sync.RWMutex
}

func NewMap() *ConcurrentMap {
	return &ConcurrentMap{
		kv: make(map[int64]*object.Process),
	}
}

func (cm *ConcurrentMap) Put(pid int64, process *object.Process) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	cm.kv[pid] = process
}

func (cm *ConcurrentMap) Get(pid int64) (*object.Process, bool) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	process, ok := cm.kv[pid]
	return process, ok
}

func (cm *ConcurrentMap) Remove(pid int64) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	delete(cm.kv, pid)
}
