package test

import (
	"reflect"
	"testing"
	"time"

	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/typesw"
)

func TestIndexHeapUpdateReordersTop(t *testing.T) {
	h := cw.NewIndexHeap[int, int](nil)
	h.Insert(1, 5)
	h.Insert(2, 1)

	if got := h.TopIndex(); got != 2 {
		t.Fatalf("unexpected initial top index: %d", got)
	}

	h.Update(1, 0)
	if got := h.TopIndex(); got != 1 {
		t.Fatalf("index heap update should reorder top, got %d", got)
	}
}

func TestLinkedListMergeKeepsOrderAndSize(t *testing.T) {
	l1 := cw.NewLinkedList[int](1, 2)
	l2 := cw.NewLinkedList[int](3, 4)

	l1.Merge(l1.Back(), l2)

	if l1.Len() != 4 {
		t.Fatalf("merged list size mismatch, got %d", l1.Len())
	}
	if !l2.Empty() {
		t.Fatal("merged source list should be cleared")
	}
	if !reflect.DeepEqual(l1.ToSlice(), []int{1, 2, 3, 4}) {
		t.Fatalf("unexpected merged order: %v", l1.ToSlice())
	}

	back := l1.PopBack()
	if back == nil || back.Value() != 4 {
		t.Fatalf("unexpected first pop back after merge: %+v", back)
	}
	back = l1.PopBack()
	if back == nil || back.Value() != 3 {
		t.Fatalf("unexpected second pop back after merge: %+v", back)
	}
}

func TestDequeToSliceAndToStringSliceDoNotBlock(t *testing.T) {
	dq := cw.NewDeque()
	dq.PushBack("a")
	dq.PushBack("b")
	dq.PushBack("c")

	doneSlice := make(chan []interface{}, 1)
	go func() {
		doneSlice <- dq.ToSlice()
	}()

	select {
	case got := <-doneSlice:
		if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
			t.Fatalf("unexpected deque slice: %+v", got)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatal("ToSlice blocked unexpectedly")
	}

	doneString := make(chan []string, 1)
	go func() {
		doneString <- dq.ToStringSlice()
	}()

	select {
	case got := <-doneString:
		if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
			t.Fatalf("unexpected deque string slice: %+v", got)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatal("ToStringSlice blocked unexpectedly")
	}
}

func TestDequeTryHelpers(t *testing.T) {
	dq := cw.NewDeque()
	if _, ok := dq.TryFront(); ok {
		t.Fatal("TryFront should return false on empty deque")
	}
	if _, ok := dq.TryBack(); ok {
		t.Fatal("TryBack should return false on empty deque")
	}
	if _, ok := dq.TryPopFront(); ok {
		t.Fatal("TryPopFront should return false on empty deque")
	}
	if _, ok := dq.TryPopBack(); ok {
		t.Fatal("TryPopBack should return false on empty deque")
	}

	dq.PushBack(1)
	if v, ok := dq.TryFront(); !ok || v.(int) != 1 {
		t.Fatalf("TryFront failed, got value=%v ok=%v", v, ok)
	}
	if v, ok := dq.TryBack(); !ok || v.(int) != 1 {
		t.Fatalf("TryBack failed, got value=%v ok=%v", v, ok)
	}
	if v, ok := dq.TryPopFront(); !ok || v.(int) != 1 {
		t.Fatalf("TryPopFront failed, got value=%v ok=%v", v, ok)
	}
	if !dq.Empty() {
		t.Fatal("deque should be empty after TryPopFront")
	}
}

func TestHeapToListKeepsAllElements(t *testing.T) {
	h := cw.NewHeap[int](nil)
	h.Insert(3)
	h.Insert(1)
	h.Insert(2)

	vals := h.ToList()
	if len(vals) != 3 {
		t.Fatalf("ToList should include all heap elements, got %d", len(vals))
	}
}

func TestStackTryPopOnEmpty(t *testing.T) {
	s := cw.NewStack[int]()
	if _, ok := s.TryPop(); ok {
		t.Fatal("TryPop should return false on empty stack")
	}
	if s.Pop() != 0 {
		t.Fatal("Pop on empty stack should return zero value")
	}

	s.Push(7)
	if v, ok := s.TryPop(); !ok || v != 7 {
		t.Fatalf("TryPop failed, got value=%v ok=%v", v, ok)
	}
}

func TestConcurrentHashMapForEachAndForEachEntryNoPanicOnNilBuckets(t *testing.T) {
	m := cw.NewConcurrentHashMap[int, int](nil, nil)
	m.Put(1, 1)
	m.Put(2, 4)

	panicCh := make(chan any, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				panicCh <- r
			}
		}()
		m.ForEach(func(k int) {})
		m.ForEachEntry(func(entry typesw.IMapEntry[int, int]) {})
	}()

	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("ConcurrentHashMap for-each operations blocked unexpectedly")
	}

	select {
	case p := <-panicCh:
		t.Fatalf("ConcurrentHashMap for-each panicked: %v", p)
	default:
	}
}

func TestTrieDeletePrefixDoesNotCorruptState(t *testing.T) {
	trie := cw.NewTrie()
	if err := trie.Insert("abc"); err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	if trie.Delete("ab") {
		t.Fatal("deleting a non-word prefix should return false")
	}
	if trie.Len() != 1 {
		t.Fatalf("trie length should remain 1, got %d", trie.Len())
	}
	if !trie.Contains("abc") {
		t.Fatal("existing word should still be present")
	}
	if !trie.Delete("abc") {
		t.Fatal("deleting existing word should return true")
	}
	if trie.Len() != 0 {
		t.Fatalf("trie length should be 0 after delete, got %d", trie.Len())
	}
}

func TestWeightedDirectedGraphDeleteEdgeKeepsOtherOutgoingEdges(t *testing.T) {
	g := cw.NewWeightedDirectedGraph[int](nil)
	g.AddEdge(1, 2, 1)
	g.AddEdge(1, 3, 2)

	if !g.DeleteEdge(1, 2) {
		t.Fatal("DeleteEdge(1,2) should succeed")
	}
	if g.Adj(1).Contains(2) {
		t.Fatal("edge 1->2 should be removed from directed graph adjacency")
	}
	if !g.Adj(1).Contains(3) {
		t.Fatal("edge 1->3 should remain after deleting 1->2")
	}

	has13 := false
	for e := range g.Edges().Iterate() {
		if e.V1() == 1 && e.V2() == 2 {
			t.Fatal("deleted edge 1->2 should not appear in weighted edge set")
		}
		if e.V1() == 1 && e.V2() == 3 {
			has13 = true
		}
	}
	if !has13 {
		t.Fatal("edge 1->3 should still exist in weighted edge set")
	}
}

func TestWeightedUndirectedGraphDeleteEdgeKeepsOtherEdges(t *testing.T) {
	g := cw.NewWeightedUndirectedGraph[int](nil)
	g.AddEdge(1, 2, 1)
	g.AddEdge(1, 3, 2)

	if !g.DeleteEdge(1, 2) {
		t.Fatal("DeleteEdge(1,2) should succeed")
	}
	if g.NumEdges() != 1 {
		t.Fatalf("expected 1 undirected edge left, got %d", g.NumEdges())
	}
	if g.Connected(1, 2) {
		t.Fatal("1 and 2 should be disconnected after deleting edge 1-2")
	}
	if !g.Connected(1, 3) {
		t.Fatal("1 and 3 should remain connected")
	}

	for e := range g.Edges().Iterate() {
		uv := e.V1() == 1 && e.V2() == 2
		vu := e.V1() == 2 && e.V2() == 1
		if uv || vu {
			t.Fatal("deleted edge 1-2 should not appear in weighted edge set")
		}
	}
}
