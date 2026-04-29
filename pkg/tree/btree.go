package tree

import "cmp"

const defaultMinDegree = 2 // minimum number of children (t)

// bnode represents a node of the B-Tree.
type bnode[K cmp.Ordered] struct {
	keys     []K
	children []*bnode[K]
	isLeaf   bool
}

func newBNode[K cmp.Ordered](t int, leaf bool) *bnode[K] {
	return &bnode[K]{
		keys:     make([]K, 0, 2*t-1),
		children: make([]*bnode[K], 0, 2*t),
		isLeaf:   leaf,
	}
}

// btree is the unexported base structure used by BTreeSet and BTreeMap.
type btree[K cmp.Ordered] struct {
	root      *bnode[K]
	minDegree int // t (minimum children = t, maximum = 2t)
	size      int
}

func newBTree[K cmp.Ordered](minDegree int) *btree[K] {
	if minDegree < 2 {
		minDegree = 2
	}
	return &btree[K]{
		root:      newBNode[K](minDegree, true),
		minDegree: minDegree,
	}
}

func (t *btree[K]) Len() int { return t.size }

// search returns the node and index if the key exists.
func (t *btree[K]) search(key K) (*bnode[K], int) {
	return t.searchNode(t.root, key)
}

func (t *btree[K]) searchNode(x *bnode[K], key K) (*bnode[K], int) {
	i := 0
	for i < len(x.keys) && key > x.keys[i] {
		i++
	}
	if i < len(x.keys) && key == x.keys[i] {
		return x, i
	}
	if x.isLeaf {
		return nil, 0
	}
	return t.searchNode(x.children[i], key)
}

// insert inserts a key; returns false if it already exists.
func (t *btree[K]) insert(key K) bool {
	if node, _ := t.search(key); node != nil {
		return false
	}
	root := t.root
	if len(root.keys) == 2*t.minDegree-1 {
		// Root is full, split it.
		newRoot := newBNode[K](t.minDegree, false)
		newRoot.children = append(newRoot.children, root)
		t.splitChild(newRoot, 0)
		t.root = newRoot
	}
	return t.insertNonFull(t.root, key)
}

func (t *btree[K]) insertNonFull(x *bnode[K], key K) bool {
	i := len(x.keys) - 1
	if x.isLeaf {
		// Detect duplicates.
		for j := 0; j < len(x.keys); j++ {
			if x.keys[j] == key {
				return false
			}
		}
		x.keys = append(x.keys, key) // placeholder slot
		for i >= 0 && key < x.keys[i] {
			x.keys[i+1] = x.keys[i]
			i--
		}
		x.keys[i+1] = key
		t.size++
		return true
	}
	for i >= 0 && key < x.keys[i] {
		i--
	}
	i++
	if len(x.children[i].keys) == 2*t.minDegree-1 {
		t.splitChild(x, i)
		if key > x.keys[i] {
			i++
		}
	}
	return t.insertNonFull(x.children[i], key)
}

func (t *btree[K]) splitChild(parent *bnode[K], i int) {
	degree := t.minDegree // do not shadow the receiver
	y := parent.children[i]
	z := newBNode[K](degree, y.isLeaf)

	// z receives the last degree-1 keys from y.
	z.keys = append(z.keys, y.keys[degree:]...)

	// The median key of y is promoted to the parent.
	median := y.keys[degree-1]

	// Insert median at position i in the parent (no prior placeholder).
	parent.keys = append(parent.keys[:i], append([]K{median}, parent.keys[i:]...)...)

	if !y.isLeaf {
		z.children = append(z.children, y.children[degree:]...)
		y.children = y.children[:degree]
	}

	// Insert z as child i+1.
	parent.children = append(parent.children[:i+1], append([]*bnode[K]{z}, parent.children[i+1:]...)...)

	// Trim y.keys to the first degree-1 keys.
	y.keys = y.keys[:degree-1]
}

// traverse walks the tree in order.
func (t *btree[K]) traverse(f func(K)) {
	t.traverseNode(t.root, f)
}

func (t *btree[K]) traverseNode(x *bnode[K], f func(K)) {
	if x == nil {
		return
	}
	for i := 0; i < len(x.keys); i++ {
		if !x.isLeaf {
			t.traverseNode(x.children[i], f)
		}
		f(x.keys[i])
	}
	if !x.isLeaf {
		t.traverseNode(x.children[len(x.keys)], f)
	}
}

// Elements returns every key in ascending order.
func (t *btree[K]) Elements() []K {
	var result []K
	t.traverse(func(k K) { result = append(result, k) })
	return result
}

// -----------------------------------------------------------
// Deletion (full implementation).
// -----------------------------------------------------------
func (t *btree[K]) delete(key K) bool {
	if t.root == nil || t.size == 0 {
		return false
	}
	found := t.deleteFromNode(t.root, key)
	if found {
		t.size--
		if len(t.root.keys) == 0 && !t.root.isLeaf {
			t.root = t.root.children[0]
		}
	}
	return found
}

func (t *btree[K]) deleteFromNode(x *bnode[K], key K) bool {
	degree := t.minDegree
	i := 0
	for i < len(x.keys) && key > x.keys[i] {
		i++
	}
	// Case 1: the key is in x.
	if i < len(x.keys) && key == x.keys[i] {
		if x.isLeaf {
			// Simple leaf removal.
			x.keys = append(x.keys[:i], x.keys[i+1:]...)
			return true
		}
		// Internal node.
		return t.deleteInternalNode(x, i)
	}
	// Case 2: key is not in x and x is a leaf => key is absent.
	if x.isLeaf {
		return false
	}
	// Case 3: keep descending.
	flag := i == len(x.keys) // the key is past the last key of x
	if len(x.children[i].keys) < degree {
		t.fill(x, i)
	}
	// After fill, child i may have been merged; re-evaluate the index.
	if flag && i > len(x.keys) {
		i--
	}
	return t.deleteFromNode(x.children[i], key)
}

func (t *btree[K]) deleteInternalNode(x *bnode[K], idx int) bool {
	degree := t.minDegree
	key := x.keys[idx]
	// If the left child has at least 'degree' keys, swap with predecessor.
	if len(x.children[idx].keys) >= degree {
		pred := t.getPred(x, idx)
		x.keys[idx] = pred
		return t.deleteFromNode(x.children[idx], pred)
	}
	// If the right child has at least 'degree' keys, swap with successor.
	if len(x.children[idx+1].keys) >= degree {
		succ := t.getSucc(x, idx)
		x.keys[idx] = succ
		return t.deleteFromNode(x.children[idx+1], succ)
	}
	// Both children have degree-1 keys: merge and remove from the merged node.
	t.merge(x, idx)
	return t.deleteFromNode(x.children[idx], key)
}

func (t *btree[K]) getPred(x *bnode[K], idx int) K {
	cur := x.children[idx]
	for !cur.isLeaf {
		cur = cur.children[len(cur.children)-1]
	}
	return cur.keys[len(cur.keys)-1]
}

func (t *btree[K]) getSucc(x *bnode[K], idx int) K {
	cur := x.children[idx+1]
	for !cur.isLeaf {
		cur = cur.children[0]
	}
	return cur.keys[0]
}

func (t *btree[K]) fill(x *bnode[K], idx int) {
	degree := t.minDegree
	if idx > 0 && len(x.children[idx-1].keys) >= degree {
		t.rotateRight(x, idx)
	} else if idx < len(x.keys) && len(x.children[idx+1].keys) >= degree {
		t.rotateLeft(x, idx)
	} else {
		if idx < len(x.keys) {
			t.merge(x, idx)
		} else {
			t.merge(x, idx-1)
		}
	}
}

func (t *btree[K]) rotateRight(x *bnode[K], idx int) {
	child := x.children[idx]
	leftSibling := x.children[idx-1]
	// Move one key from the parent down into child.
	child.keys = append([]K{x.keys[idx-1]}, child.keys...)
	x.keys[idx-1] = leftSibling.keys[len(leftSibling.keys)-1]
	if !child.isLeaf {
		child.children = append([]*bnode[K]{leftSibling.children[len(leftSibling.children)-1]}, child.children...)
		leftSibling.children = leftSibling.children[:len(leftSibling.children)-1]
	}
	leftSibling.keys = leftSibling.keys[:len(leftSibling.keys)-1]
}

func (t *btree[K]) rotateLeft(x *bnode[K], idx int) {
	child := x.children[idx]
	rightSibling := x.children[idx+1]
	child.keys = append(child.keys, x.keys[idx])
	x.keys[idx] = rightSibling.keys[0]
	if !child.isLeaf {
		child.children = append(child.children, rightSibling.children[0])
		rightSibling.children = rightSibling.children[1:]
	}
	rightSibling.keys = rightSibling.keys[1:]
}

func (t *btree[K]) merge(x *bnode[K], idx int) {
	child := x.children[idx]
	sibling := x.children[idx+1]
	// Pull the median key from the parent down into child.
	child.keys = append(child.keys, x.keys[idx])
	child.keys = append(child.keys, sibling.keys...)
	if !child.isLeaf {
		child.children = append(child.children, sibling.children...)
	}
	// Remove the key and the right child from the parent.
	x.keys = append(x.keys[:idx], x.keys[idx+1:]...)
	x.children = append(x.children[:idx+1], x.children[idx+2:]...)
}

// Min and Max.
func (t *btree[K]) Min() (K, bool) {
	if t.size == 0 {
		var zero K
		return zero, false
	}
	cur := t.root
	for !cur.isLeaf {
		cur = cur.children[0]
	}
	return cur.keys[0], true
}

func (t *btree[K]) Max() (K, bool) {
	if t.size == 0 {
		var zero K
		return zero, false
	}
	cur := t.root
	for !cur.isLeaf {
		cur = cur.children[len(cur.children)-1]
	}
	return cur.keys[len(cur.keys)-1], true
}
