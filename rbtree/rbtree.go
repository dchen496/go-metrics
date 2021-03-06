package rbtree

// rbtree implements a red-black tree. It can be used as a multimap.
// It additionally allows finding nodes by rank.
type Tree struct {
	root      *Node
	none      *Node
	freeNodes []*Node
}

type Node struct {
	black  bool
	size   uint64
	parent *Node
	left   *Node
	right  *Node
	key    int64
	value  interface{}
}

func New() *Tree {
	t := Tree{}
	t.freeNodes = make([]*Node, 0, freeNodeCacheCap)
	n := t.newNode()
	*n = Node{black: true}
	n.parent = n
	n.left = n
	n.right = n
	t.root = n
	t.none = n
	return &t
}

// lower demand on allocator & GC: cache free nodes
const freeNodeCacheCap = 10

func (t *Tree) newNode() *Node {
	if len(t.freeNodes) > 0 {
		r := t.freeNodes[len(t.freeNodes)-1]
		t.freeNodes = t.freeNodes[0 : len(t.freeNodes)-1]
		return r
	}
	return &Node{}
}

func (t *Tree) freeNode(n *Node) {
	*n = Node{}
	if len(t.freeNodes) >= cap(t.freeNodes) {
		return
	}
	t.freeNodes = append(t.freeNodes, n)
}

func (t *Tree) Size() uint64 {
	return t.root.size
}

// FindByRank returns the node with the specified rank,
// or nil if that node does not exist.
func (t *Tree) FindByRank(rank uint64) *Node {
	if rank >= t.root.size {
		return nil
	}
	n := t.root
	for n != t.none {
		if rank == n.left.size {
			return n
		}
		if rank < n.left.size {
			n = n.left
		} else {
			rank -= n.left.size + 1
			n = n.right
		}
	}
	//should never reach here
	return nil
}

// Find returns the node with the specified key. If multiple nodes
// have the same key, the one with the lowest rank is returned.
// Find returns nil if no node with the specified key exists.
func (t *Tree) Find(key int64) *Node {
	n := t.LowerBound(key)
	if n == nil {
		return nil
	}
	if n.key != key {
		return nil
	}
	return n
}

// LowerBound finds the node with the smallest key greater than
// or equal to the argument. If multiple nodes have the same
// key, the one with the lowest rank is returned.
func (t *Tree) LowerBound(key int64) *Node {
	var result *Node
	for n := t.root; n != t.none; {
		if n.key < key {
			n = n.right
		} else {
			result = n
			n = n.left
		}
	}
	if result == nil {
		return nil
	}
	return result
}

// UpperBound finds the node with the largest key
// less than the argument. If multiple nodes have the same key,
// the one with the highest rank is returned.
func (t *Tree) UpperBound(key int64) *Node {
	var result *Node
	for n := t.root; n != t.none; {
		if n.key >= key {
			n = n.left
		} else {
			result = n
			n = n.right
		}
	}
	if result == nil {
		return nil
	}
	return result
}

func (t *Tree) Insert(key int64, value interface{}) *Node {
	prev := t.none
	for n := t.root; n != t.none; {
		prev = n
		if key < n.key {
			n = n.left
		} else {
			n = n.right
		}
	}

	n := t.newNode()
	*n = Node{false, 0, prev, t.none, t.none, key, value}
	if prev == t.none {
		t.root = n
	} else if key < prev.key {
		prev.left = n
	} else {
		prev.right = n
	}
	t.updateSize(n)
	t.insertFix(n)
	return n
}

func (t *Tree) subTreeMin(n *Node) *Node {
	for n.left != t.none {
		n = n.left
	}
	return n
}

func (t *Tree) subTreeMax(n *Node) *Node {
	for n.right != t.none {
		n = n.right
	}
	return n
}

func (t *Tree) rotateLeft(n *Node) {
	r := n.right
	if r == t.none {
		return
	}

	n.right = r.left
	if n.right != t.none {
		n.right.parent = n
	}

	p := n.parent
	if p == t.none {
		t.root = r
	} else if n == p.left {
		p.left = r
	} else {
		p.right = r
	}
	r.parent = p

	r.left = n
	n.parent = r

	t.updateSize(n)
	t.updateSize(r)
}

func (t *Tree) rotateRight(n *Node) {
	l := n.left
	if l == t.none {
		return
	}

	n.left = l.right
	if n.left != t.none {
		n.left.parent = n
	}

	p := n.parent
	if p == t.none {
		t.root = l
	} else if n == p.right {
		p.right = l
	} else {
		p.left = l
	}
	l.parent = p

	l.right = n
	n.parent = l

	t.updateSize(n)
	t.updateSize(l)
}

func (t *Tree) insertFix(n *Node) {
	for n.parent.black == false {
		var uncle *Node
		if n.parent == n.parent.parent.left {
			uncle = n.parent.parent.right
			if uncle.black == false {
				n.parent.black = true
				uncle.black = true
				n.parent.parent.black = false
				n = n.parent.parent
			} else {
				if n == n.parent.right {
					n = n.parent
					t.rotateLeft(n)
				}
				n.parent.black = true
				n.parent.parent.black = false
				t.rotateRight(n.parent.parent)
			}
		} else { //symmetric to above
			uncle = n.parent.parent.left
			if uncle.black == false {
				n.parent.black = true
				uncle.black = true
				n.parent.parent.black = false
				n = n.parent.parent
			} else {
				if n == n.parent.left {
					n = n.parent
					t.rotateRight(n)
				}
				n.parent.black = true
				n.parent.parent.black = false
				t.rotateLeft(n.parent.parent)
			}
		}
	}
	t.root.black = true
}

func (t *Tree) Remove(key int64) {
	n := t.Find(key)
	if n == nil {
		return
	}
	t.RemoveNode(n)
}

// Remember to clear pointers to the removed node.
func (t *Tree) RemoveNode(n *Node) {
	y := n
	yBlack := y.black
	var x *Node
	if n.left == t.none {
		x = n.right
		t.transplant(n, x)
	} else if n.right == t.none {
		x = n.left
		t.transplant(n, x)
	} else {
		y = t.subTreeMin(n.right)
		yBlack = y.black
		x = y.right
		if y.parent == n {
			x.parent = y // in case x is t.none
		} else {
			t.transplant(y, x)
			y.right = n.right
			y.right.parent = y
		}
		t.transplant(n, y)
		y.left = n.left
		t.updateSize(y)
		y.left.parent = y
		y.black = n.black
	}
	if yBlack == true {
		t.removeFix(x)
	}
	t.none.parent = t.none
	t.freeNode(n)
}

func (t *Tree) updateSize(n *Node) {
	for ; n != t.none; n = n.parent {
		newSize := n.left.size + n.right.size + 1
		if n.size == newSize {
			return
		}
		n.size = newSize
	}
}

func (t *Tree) transplant(dst *Node, src *Node) {
	newp := dst.parent
	if newp == t.none {
		t.root = src
	} else if dst == newp.left {
		newp.left = src
	} else {
		newp.right = src
	}
	t.updateSize(newp)
	src.parent = newp
}

func (t *Tree) removeFix(n *Node) {
	for n != t.root && n.black == true {
		var w *Node
		if n == n.parent.left {
			w = n.parent.right
			if w.black == false {
				w.black = true
				n.parent.black = false
				t.rotateLeft(n.parent)
				w = n.parent.right
			}
			if w.left.black == true && w.right.black == true {
				w.black = false
				n = n.parent
			} else {
				if w.right.black == true {
					w.left.black = true
					w.black = false
					t.rotateRight(w)
					w = n.parent.right
				}
				w.black = n.parent.black
				n.parent.black = true
				w.right.black = true
				t.rotateLeft(n.parent)
				n = t.root
			}
		} else { //symmetric to above
			w = n.parent.left
			if w.black == false {
				w.black = true
				n.parent.black = false
				t.rotateRight(n.parent)
				w = n.parent.left
			}
			if w.right.black == true && w.left.black == true {
				w.black = false
				n = n.parent
			} else {
				if w.left.black == true {
					w.right.black = true
					w.black = false
					t.rotateLeft(w)
					w = n.parent.left
				}
				w.black = n.parent.black
				n.parent.black = true
				w.left.black = true
				t.rotateRight(n.parent)
				n = t.root
			}
		}
	}
	n.black = true
}

// Next returns the node with rank one greater than the
// current one, or nil if the current node is the last node.
func (t *Tree) Next(n *Node) *Node {
	if n == t.none {
		return nil
	}
	if n.right != t.none {
		return t.subTreeMin(n.right)
	}
	for n == n.parent.right {
		n = n.parent
		if n == t.none {
			return nil
		}
	}
	if n.parent == t.none {
		return nil
	}
	return n.parent
}

// Prev returns the node with rank one less than the
// current one, or nil if the current node is the first node.
func (t *Tree) Prev(n *Node) *Node {
	if n == t.none {
		return nil
	}
	if n.left != t.none {
		return t.subTreeMax(n.left)
	}
	for n == n.parent.left {
		n = n.parent
		if n == t.none {
			return nil
		}
	}
	if n.parent == t.none {
		return nil
	}
	return n.parent
}

// Rank returns the rank of its argument, or the number of nodes
// with keys less than than that of the argument, plus a
// nonnegative integer less than the number of nodes with the same key.
// Any two distinct nodes have a different rank, even if they share
// the same key.
func (t *Tree) Rank(n *Node) uint64 {
	rank := n.left.size
	for ; n != t.root; n = n.parent {
		if n == n.parent.right {
			rank += 1 + n.parent.left.size
		}
	}
	return rank
}

func (n *Node) Key() int64 {
	return n.key
}

func (n *Node) Value() interface{} {
	return n.value
}

/* untested
func (t *Tree) Iter(n *Node) chan *Node {
  ch := make(chan *Node)
  go func(){
    for n != nil {
      ch <- n
      n = t.Next(n)
    }
  }()
  return ch
}

func (t *Tree) RevIter(n *Node) chan *Node {
  ch := make(chan *Node)
  go func(){
    for n != nil {
      ch <- n
      n = t.Prev(n)
    }
  }()
  return ch
}
*/
