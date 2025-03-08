package object

// Struct
type OrderedMap2[K comparable, V any] struct {
	store map[K]V
	Keys  []K
}

func NewPairsMap() OrderedMap2[HashKey, MapPair] {
	return *NewOrderedMap[HashKey, MapPair]()
}

func NewPairsMapWithSize(size int) OrderedMap2[HashKey, MapPair] {
	return *NewOrderedMapWithSize[HashKey, MapPair](size)
}

// Constructor
func NewOrderedMap[K comparable, V any]() *OrderedMap2[K, V] {
	return &OrderedMap2[K, V]{
		store: map[K]V{},
		Keys:  []K{},
	}
}

func NewOrderedMapWithSize[K comparable, V any](size int) *OrderedMap2[K, V] {
	return &OrderedMap2[K, V]{
		store: make(map[K]V, size),
		Keys:  make([]K, 0, size),
	}
}

// Get will return the value associated with the key.
// If the key doesn't exist, the second return value will be false.
func (o *OrderedMap2[K, V]) Get(key K) (V, bool) {
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
	if _, exists := o.store[key]; !exists {
		o.Keys = append(o.Keys, key)
	}
	o.store[key] = val
}

// Delete will remove the key and its associated value.
func (o *OrderedMap2[K, V]) Delete(key K) {
	delete(o.store, key)

	// Find key in slice
	idx := -1

	for i, val := range o.Keys {
		if val == key {
			idx = i
			break
		}
	}
	if idx != -1 {
		o.Keys = append(o.Keys[:idx], o.Keys[idx+1:]...)
	}
}
