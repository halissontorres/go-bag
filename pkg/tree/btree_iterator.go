package tree

import "cmp"

// BTreeIterator iterates over keys in ascending order.
type BTreeIterator[K cmp.Ordered] struct {
	tree     *btree[K]
	node     *bnode[K]
	idx      int
	started  bool
	finished bool
}

// newBTreeIterator creates an iterator positioned on the smallest element.
func newBTreeIterator[K cmp.Ordered](tree *btree[K]) *BTreeIterator[K] {
	it := &BTreeIterator[K]{tree: tree}
	if tree.size == 0 {
		it.finished = true
		return it
	}
	// Position on the smallest element.
	it.node = tree.root
	for !it.node.isLeaf {
		it.node = it.node.children[0]
	}
	it.idx = 0
	it.started = true
	return it
}

// newBTreeIteratorFromKey creates an iterator positioned on the first element >= key.
func newBTreeIteratorFromKey[K cmp.Ordered](tree *btree[K], key K) *BTreeIterator[K] {
	it := &BTreeIterator[K]{tree: tree}
	if tree.size == 0 {
		it.finished = true
		return it
	}
	node := tree.root
	for {
		i := 0
		for i < len(node.keys) && key > node.keys[i] {
			i++
		}
		if i < len(node.keys) && node.keys[i] >= key {
			// We may still find a smaller-but-valid key in the left subtree;
			// descend to the leaf while preserving the position.
			if node.isLeaf {
				it.node = node
				it.idx = i
				it.started = true
				return it
			}
			// Descend into child i.
			node = node.children[i]
		} else {
			// key is greater than every key in this node; descend into the last child.
			if node.isLeaf {
				it.finished = true
				return it
			}
			node = node.children[len(node.children)-1]
		}
	}
}

// Next returns the next key and advances the iterator. It returns false when exhausted.
func (it *BTreeIterator[K]) Next() (K, bool) {
	if it.finished || !it.started || it.node == nil {
		var zero K
		return zero, false
	}
	// Capture the current value.
	val := it.node.keys[it.idx]

	// Advance to the next element.
	if !it.node.isLeaf {
		// With a right subtree, the successor is the minimum of that subtree.
		child := it.node.children[it.idx+1]
		// Walk all the way to the leftmost leaf.
		for !child.isLeaf {
			child = child.children[0]
		}
		it.node = child
		it.idx = 0
	} else {
		// Leaf: move to the next index.
		it.idx++
		// If we ran past the end, climb until we find a parent where we sit on the left.
		for it.idx >= len(it.node.keys) {
			if it.node == it.tree.root {
				it.finished = true
				return val, true
			}
			// Find the parent and the index at which it.node sits among its children.
			parent := it.tree.findParent(it.tree.root, it.node)
			if parent == nil {
				it.finished = true
				return val, true
			}
			childIdx := -1
			for i, c := range parent.children {
				if c == it.node {
					childIdx = i
					break
				}
			}
			if childIdx < len(parent.keys) {
				// The parent key at childIdx is the next element.
				it.node = parent
				it.idx = childIdx
				break
			} else {
				// No parent key applies; keep climbing.
				it.node = parent
				it.idx = len(it.node.keys) // force another climb
			}
		}
	}
	return val, true
}

// Prev would mirror Next descending leftward; it is intentionally omitted.
// Range queries are exposed through Range below, which returns a slice.

// Range returns a slice with all keys in the closed interval [low, high].
func (t *btree[K]) Range(low, high K) []K {
	result := []K{}
	t.rangeTraverse(t.root, low, high, &result)
	return result
}

func (t *btree[K]) rangeTraverse(x *bnode[K], low, high K, result *[]K) {
	if x == nil {
		return
	}
	i := 0
	for i < len(x.keys) && x.keys[i] < low {
		i++
	}
	for i < len(x.keys) && x.keys[i] <= high {
		if !x.isLeaf {
			t.rangeTraverse(x.children[i], low, high, result)
		}
		*result = append(*result, x.keys[i])
		i++
	}
	if !x.isLeaf {
		t.rangeTraverse(x.children[i], low, high, result)
	}
}

// findParent is an internal helper used for navigation.
func (t *btree[K]) findParent(current, child *bnode[K]) *bnode[K] {
	if current == nil || current.isLeaf {
		return nil
	}
	for i := 0; i <= len(current.keys); i++ {
		if current.children[i] == child {
			return current
		}
		if i <= len(current.keys) {
			if res := t.findParent(current.children[i], child); res != nil {
				return res
			}
		}
	}
	return nil
}
