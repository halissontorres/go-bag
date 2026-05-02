package tree

import "cmp"

// BTreeMap is a key-ordered map backed by a B-Tree.
type BTreeMap[K cmp.Ordered, V any] struct {
	tree *btree[K]
	vals map[K]V
}

// KeyValuePair holds a key and its associated value.
type KeyValuePair[K any, V any] struct {
	Key   K
	Value V
}

func NewBTreeMap[K cmp.Ordered, V any](opts ...Option) *BTreeMap[K, V] {
	c := applyTreeOptions(opts)
	return &BTreeMap[K, V]{
		tree: newBTree[K](c.minDegree),
		vals: make(map[K]V),
	}
}

func (m *BTreeMap[K, V]) Put(key K, value V) {
	if m.tree.insert(key) {
		m.vals[key] = value
	} else {
		m.vals[key] = value
	}
}

func (m *BTreeMap[K, V]) Get(key K) (V, bool) {
	if node, _ := m.tree.search(key); node != nil {
		return m.vals[key], true
	}
	var zero V
	return zero, false
}

func (m *BTreeMap[K, V]) Remove(key K) bool {
	if !m.Contains(key) {
		return false
	}
	delete(m.vals, key)
	return m.tree.delete(key)
}

func (m *BTreeMap[K, V]) Contains(key K) bool {
	node, _ := m.tree.search(key)
	return node != nil
}

func (m *BTreeMap[K, V]) Len() int { return m.tree.Len() }

func (m *BTreeMap[K, V]) ForEach(f func(K, V)) {
	m.tree.traverse(func(key K) {
		f(key, m.vals[key])
	})
}

func (m *BTreeMap[K, V]) Keys() []K {
	return m.tree.Elements()
}

func (m *BTreeMap[K, V]) Values() []V {
	var vals []V
	for _, k := range m.Keys() {
		vals = append(vals, m.vals[k])
	}
	return vals
}

// Min returns the smallest key and its value. Runs in O(log n).
func (m *BTreeMap[K, V]) Min() (K, V, bool) {
	key, ok := m.tree.Min()
	if !ok {
		var v V
		return key, v, false
	}
	val, _ := m.Get(key)
	return key, val, true
}

// Max returns the largest key and its value. Runs in O(log n).
func (m *BTreeMap[K, V]) Max() (K, V, bool) {
	key, ok := m.tree.Max()
	if !ok {
		var v V
		return key, v, false
	}
	val, _ := m.Get(key)
	return key, val, true
}

// Range returns the key-value pairs whose keys lie in the closed interval [low, high].
func (m *BTreeMap[K, V]) Range(low, high K) []KeyValuePair[K, V] {
	keys := m.tree.Range(low, high)
	pairs := make([]KeyValuePair[K, V], len(keys))
	for i, k := range keys {
		v, _ := m.Get(k)
		pairs[i] = KeyValuePair[K, V]{Key: k, Value: v}
	}
	return pairs
}
