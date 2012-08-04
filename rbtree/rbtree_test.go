package rbtree

import (
	"fmt"
	"math/rand"
	"testing"
)

const (
	testRandSeed          = 1234
	testNumElements       = 1000
	testBenchmarkElements = 1000000
)

func testInit() *Tree {
	rb := New()
	rand.Seed(testRandSeed)
	p := rand.Perm(testNumElements)
	for _, v := range p {
		rb.Insert(int64(v)*10, fmt.Sprint(v*10))
	}
	return rb
}

func testTreeStructure(t *testing.T, rb *Tree, data []int, removed *Node) {
	needs := make(map[int]bool)
	for _, v := range data {
		needs[v*10] = true
	}

	if rb.root.black == false {
		t.Errorf("Tree root is not black\n")
	}

	var blackCount uint64
	for p := rb.FindByRank(0); p != rb.none; p = p.parent {
		if p.black {
			blackCount++
		}
	}

	for n := rb.FindByRank(0); n != nil; n = rb.Next(n) {
		if n == rb.none {
			t.Errorf("Unexpected sentinel")
		}
		if n == removed || n.left == removed || n.right == removed || n.parent == removed {
			t.Errorf("Unexpected node")
		}
		if needs[int(n.key)] == false && data != nil {
			t.Errorf("Unexpected key %d found in tree", n.key)
		}
		needs[int(n.key)] = false

		if n.size != n.left.size+n.right.size+1 {
			t.Errorf("Wrong size for node %d: Size: %d Left: %d Right: %d\n",
				n.key, n.size, n.left.size, n.right.size)
		}

		if n.left == rb.none && n.right == rb.none {
			var nodeBlackCount uint64
			for p := n; p != rb.none; p = p.parent {
				if p.black {
					nodeBlackCount++
				}
			}
			if nodeBlackCount != blackCount {
				t.Errorf("Node %d violates black-length invariant: %d vs %d\n",
					n.key, nodeBlackCount, blackCount)
			}
		}

		if n.black == false && (n.left.black == false || n.right.black == false) {
			t.Errorf("Children of red node %d are not black\n", n.key)
		}

		if n.parent != rb.none && n.parent.left != n && n.parent.right != n {
			t.Errorf("Node %d is not correctly linked to its parent %d\n",
				n.key, n.parent.key)
		}

		if rb.root.black != true {
			t.Errorf("Tree root is not black.")
		}

		if n.left != rb.none && n.left.key > n.key {
			t.Errorf("Tree is not in the right order")
		}

		if n.right != rb.none && n.right.key < n.key {
			t.Errorf("Tree is not in the right order")
		}
	}

	for _, v := range data {
		if needs[v*10] == true {
			t.Errorf("Key %d not found in tree", v)
		}
	}
}

func testKeyValue(t *testing.T, n *Node, value int64) {
	if n == nil {
		t.Errorf("Node %d doesn't exist.\n", value)
	}
	if n.Key() != int64(value) {
		t.Errorf("Wrong node for key %d: got key %d\n", value, n.Key())
	}
	if n.Value().(string) != fmt.Sprint(value) {
		t.Errorf("Wrong value for key %d: got value %s\n",
			value, n.Value().(string))
	}
}

func TestInsert(t *testing.T) {
	rb := New()
	rand.Seed(testRandSeed)
	p := rand.Perm(testNumElements)
	for i, v := range p {
		rb.Insert(int64(v)*10, fmt.Sprint(v*10))
		testTreeStructure(t, rb, p[:i+1], nil)
	}
}

func TestDuplicateInsert(t *testing.T) {
	rb := New()
	rand.Seed(testRandSeed)
	for i := 0; i < 10; i++ {
		p := rand.Perm(testNumElements)
		for _, v := range p {
			rb.Insert(int64(v)*10, fmt.Sprint(v*10))
			testTreeStructure(t, rb, nil, nil)
		}
	}
}

func TestRemove(t *testing.T) {
	rb := testInit()
	rand.Seed(testRandSeed * 2)
	p := rand.Perm(testNumElements)

	testTreeStructure(t, rb, p, nil)

	for i, v := range p {
		n := rb.Find(int64(v) * 10)
		rb.Remove(int64(v) * 10)
		if rb.Size() != 0 {
			testTreeStructure(t, rb, p[i+1:], n)
		}
	}
	if rb.root != rb.none {
		t.Errorf("Root is not none after removing all nodes\n")
	}
}

func TestFind(t *testing.T) {
	rb := testInit()
	for i := 0; i < testNumElements; i++ {
		key := int64(i) * 10
		n := rb.Find(key)
		testKeyValue(t, n, key)
	}
}

func TestFindByRank(t *testing.T) {
	rb := testInit()
	for i := 0; i < testNumElements; i++ {
		key := int64(i) * 10
		n := rb.FindByRank(uint64(i))
		testKeyValue(t, n, key)
	}
}

func TestLowerBound(t *testing.T) {
	rb := testInit()
	for i := 0; i < testNumElements; i++ {
		key := int64(i)*10 - 5
		n := rb.LowerBound(key)
		testKeyValue(t, n, key+5)
	}
	for i := 0; i < testNumElements; i++ {
		key := int64(i) * 10
		n := rb.LowerBound(key)
		testKeyValue(t, n, key)
	}
	n := rb.LowerBound(int64(testNumElements * 10))
	if n != nil {
		t.Errorf("LowerBound should be nil beyond last element, got %v\n", n)
	}
}

func TestUpperBound(t *testing.T) {
	rb := testInit()
	for i := 0; i < testNumElements; i++ {
		key := int64(i)*10 + 5
		n := rb.UpperBound(key)
		testKeyValue(t, n, key-5)
	}
	for i := 1; i < testNumElements+1; i++ {
		key := int64(i) * 10
		n := rb.UpperBound(key)
		testKeyValue(t, n, key-10)
	}
	n := rb.UpperBound(int64(0))
	if n != nil {
		t.Errorf("UpperBound should be nil before first element, got %v\n", n)
	}
}

func TestNext(t *testing.T) {
	rb := testInit()
	n := rb.FindByRank(0)
	for i := 0; i < testNumElements; i++ {
		key := int64(i) * 10
		testKeyValue(t, n, key)
		n = rb.Next(n)
	}
}

func TestPrev(t *testing.T) {
	rb := testInit()
	n := rb.FindByRank(rb.Size() - 1)
	for i := testNumElements - 1; i >= 0; i-- {
		key := int64(i) * 10
		testKeyValue(t, n, key)
		n = rb.Prev(n)
	}
}

func TestRank(t *testing.T) {
	rb := testInit()
	for i := uint64(0); i < testNumElements; i++ {
		n := rb.FindByRank(i)
		if rb.Rank(n) != i {
			t.Errorf("Node rank should be %d, got %d", i, rb.Rank(n))
		}
	}
}

func BenchmarkInsert(b *testing.B) {
	rb := New()
	rand.Seed(testRandSeed)
	for i := 0; i < b.N; i++ {
		rb.Insert(rand.Int63(), nil)
		if i%testBenchmarkElements == 0 {
			rb = New()
		}
	}
}

func BenchmarkTransverse(b *testing.B) {
	b.StopTimer()
	rb := testInit()
	n := rb.FindByRank(0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		n = rb.Next(n)
		if n == nil {
			n = rb.FindByRank(0)
		}
	}
}

func BenchmarkSearch(b *testing.B) {
	b.StopTimer()
	rb := New()
	rand.Seed(testRandSeed)
	for i := 0; i < testBenchmarkElements; i++ {
		rb.Insert(rand.Int63(), nil)
	}
	b.StartTimer()
	sum := int64(0)
	for i := 0; i < b.N; i++ {
		if i%testBenchmarkElements == 0 {
			rand.Seed(testRandSeed)
		}
		n := rb.Find(rand.Int63())
		sum += n.Key()
	}
}
