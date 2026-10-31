package object

const (
	smallMapThreshold   = 8
	pairsMapDefaultSize = 4
)

type OrderedMap2[K comparable, V any] struct {
	Keys  []K
	vals  []V
	store map[K]V
}

func NewPairsMap() OrderedMap2[HashKey, MapPair] {
	return *NewOrderedMapWithSize[HashKey, MapPair](pairsMapDefaultSize)
}

func NewPairsMapWithSize(size int) OrderedMap2[HashKey, MapPair] {
	return *NewOrderedMapWithSize[HashKey, MapPair](size)
}

func NewOrderedMap[K comparable, V any]() *OrderedMap2[K, V] {
	return &OrderedMap2[K, V]{}
}

func NewOrderedMapWithSize[K comparable, V any](size int) *OrderedMap2[K, V] {
	if size > smallMapThreshold {
		return &OrderedMap2[K, V]{
			store: make(map[K]V, size),
			Keys:  make([]K, 0, size),
		}
	}
	return &OrderedMap2[K, V]{
		Keys: make([]K, 0, size),
		vals: make([]V, 0, size),
	}
}

func (o *OrderedMap2[K, V]) Small() bool {
	return o.store == nil
}

// Get will return the value associated with the key.
// If the key doesn't exist, the second return value will be false.
func (o *OrderedMap2[K, V]) Get(key K) (V, bool) {
	if o.store == nil {
		for i, k := range o.Keys {
			if k == key {
				return o.vals[i], true
			}
		}
		var zero V
		return zero, false
	}
	val, exists := o.store[key]
	return val, exists
}

// Len returns the number of keys in the map
func (o *OrderedMap2[K, V]) Len() int {
	return len(o.Keys)
}

// Set will store a key-value pair. If the key already exists,
// it will overwrite the existing key-value pair.
func (o *OrderedMap2[K, V]) Set(key K, val V) {
	if o.store == nil {
		for i, k := range o.Keys {
			if k == key {
				o.vals[i] = val
				return
			}
		}
		o.appendSmall(key, val)
		return
	}
	if _, exists := o.store[key]; !exists {
		o.Keys = append(o.Keys, key)
	}
	o.store[key] = val
}

func (o *OrderedMap2[K, V]) appendSmall(key K, val V) {
	o.Keys = append(o.Keys, key)
	o.vals = append(o.vals, val)
	if len(o.Keys) > smallMapThreshold {
		o.promote()
	}
}

func (o *OrderedMap2[K, V]) promote() {
	store := make(map[K]V, len(o.Keys))
	for i, k := range o.Keys {
		store[k] = o.vals[i]
	}
	o.store = store
	o.vals = nil
}

// Delete will remove the key and its associated value.
func (o *OrderedMap2[K, V]) Delete(key K) {
	if o.store == nil {
		for i, k := range o.Keys {
			if k == key {
				o.removeAt(i)
				return
			}
		}
		return
	}
	delete(o.store, key)
	for i, k := range o.Keys {
		if k == key {
			o.removeAt(i)
			return
		}
	}
}

func (o *OrderedMap2[K, V]) removeAt(i int) {
	o.Keys = append(o.Keys[:i], o.Keys[i+1:]...)
	if o.store == nil {
		o.vals = append(o.vals[:i], o.vals[i+1:]...)
	}
}
