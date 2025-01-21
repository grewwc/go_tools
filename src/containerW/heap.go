package containerW

import (
	"github.com/grewwc/go_tools/src/containerW/typesW"
	"golang.org/x/exp/constraints"
)

type heapComparable struct {
	data []interface{}
}

type heap[T constraints.Ordered] struct {
	data []T
}

type Heap[T constraints.Ordered] interface {
	Insert(T)
	Pop() T
	Size() int
	IsEmpty() bool
	ToList() []T
	Next() T
}

type HeapComparable interface {
	Insert(interface{})
	Pop() interface{}
	Size() int
	IsEmpty() bool
	ToList() []interface{}
	Next() interface{}
}

type MinHeapComparable heapComparable

type MaxHeapComparable heapComparable

type MinHeap[T constraints.Ordered] heap[T]

type MaxHeap[T constraints.Ordered] heap[T]

func newHeapCapComparable(cap int) *heapComparable {
	return &heapComparable{data: make([]interface{}, 1, cap+1)}
}

func newHeapCap[T constraints.Ordered](cap int) *heap[T] {
	return &heap[T]{data: make([]T, 1, cap+1)}
}

func (h *heapComparable) size() int {
	return len(h.data) - 1
}

func (h *heap[T]) size() int {
	return len(h.data) - 1
}

func (h *heapComparable) insert(val interface{}, reverse bool) {
	h.data = append(h.data, val)
	swimComparable(h.data, len(h.data)-1, reverse)
}

func (h *heap[T]) insert(val T, reverse bool) {
	h.data = append(h.data, val)
	swim(h.data, len(h.data)-1, reverse)
}

func (h *heapComparable) next() interface{} {
	if len(h.data) == 1 {
		return nil
	}
	return h.data[1]
}

func (h *heap[T]) next() T {
	if len(h.data) == 1 {
		return *new(T)
	}
	return h.data[1]
}

func (h *heapComparable) tolist() []interface{} {
	res := make([]interface{}, h.size()-1)
	copy(res, h.data[1:])
	return res
}

func (h *heap[T]) tolist() []T {
	res := make([]T, h.size()-1)
	copy(res, h.data[1:])
	return res
}

func (h *heapComparable) pop(reverse bool) interface{} {
	if h.size() <= 0 {
		return nil
	}
	res := h.next()
	// swap
	h.data[1], h.data[len(h.data)-1] = h.data[len(h.data)-1], h.data[1]
	h.data = h.data[:len(h.data)-1]
	sinkComparable(h.data, 1, reverse)
	return res
}

func (h *heap[T]) pop(reverse bool) T {
	if h.size() <= 0 {
		return *new(T)
	}
	res := h.next()
	// swap
	h.data[1], h.data[len(h.data)-1] = h.data[len(h.data)-1], h.data[1]
	h.data = h.data[:len(h.data)-1]
	sink(h.data, 1, reverse)
	return res
}

func (h *heapComparable) isEmpty() bool {
	return len(h.data) == 1
}

func (h *heap[T]) isEmpty() bool {
	return len(h.data) == 1
}

func larger[T constraints.Ordered](a, b T, reverse bool) bool {
	if !reverse {
		return a > b
	}
	return a < b
}

func swim[T constraints.Ordered](arr []T, idx int, reverse bool) {
	if idx <= 1 {
		return
	}

	if larger(arr[idx/2], arr[idx], reverse) {
		arr[idx/2], arr[idx] = arr[idx], arr[idx/2]
		swim(arr, idx/2, reverse)
		sink(arr, idx, reverse)
	}
}

func sink[T constraints.Ordered](arr []T, idx int, reverse bool) {
	childIdx := idx * 2
	if childIdx >= len(arr) {
		return
	}
	if childIdx+1 < len(arr) && larger(arr[childIdx], arr[childIdx+1], reverse) {
		childIdx++
	}
	if larger(arr[idx], arr[childIdx], reverse) {
		arr[childIdx], arr[idx] = arr[idx], arr[childIdx]
		sink(arr, childIdx, reverse)
	}
}

func swimComparable(arr []interface{}, idx int, reverse bool) {
	if idx <= 1 {
		return
	}
	cmp := arr[idx/2].(typesW.Comparable).Compare(arr[idx])
	if reverse {
		cmp *= -1
	}
	if cmp > 0 {
		arr[idx/2], arr[idx] = arr[idx], arr[idx/2]
		swimComparable(arr, idx/2, reverse)
		sinkComparable(arr, idx, reverse)
	}
}

func sinkComparable(arr []interface{}, idx int, reverse bool) {
	leftChild := 2 * idx
	if leftChild+1 > len(arr) {
		return
	}
	minChild := leftChild
	if leftChild+1 < len(arr) {
		cmp := arr[leftChild+1].(typesW.Comparable).Compare(arr[leftChild])
		if reverse {
			cmp *= -1
		}
		if cmp < 0 {
			minChild++
		}
	}
	cmp := arr[minChild].(typesW.Comparable).Compare(arr[idx])
	if reverse {
		cmp *= -1
	}
	if cmp < 0 {
		arr[minChild], arr[idx] = arr[idx], arr[minChild]
		sinkComparable(arr, minChild, reverse)
	}
}

// public methods
// NewMinHeap

func NewMinHeapComparable() *MinHeapComparable {
	return (*MinHeapComparable)(newHeapCapComparable(0))
}

func NewMinHeapCapComparable(cap int) *MinHeapComparable {
	return (*MinHeapComparable)(newHeapCapComparable(cap))
}

func NewMinHeap[T constraints.Ordered]() *MinHeap[T] {
	return (*MinHeap[T])(newHeapCap[T](0))
}

func NewMinHeapCap[T constraints.Ordered](cap int) *MinHeap[T] {
	return (*MinHeap[T])(newHeapCap[T](cap))
}

func (h *MinHeapComparable) Insert(val interface{}) {
	(*heapComparable)(h).insert(val, false)
}

func (h *MinHeap[T]) Insert(val T) {
	(*heap[T])(h).insert(val, false)
}

func (h *MinHeapComparable) Pop() interface{} {
	return (*heapComparable)(h).pop(false)
}

func (h *MinHeap[T]) Pop() T {
	return (*heap[T])(h).pop(false)
}

func (h *MinHeapComparable) Size() int {
	return (*heapComparable)(h).size()
}

func (h *MinHeap[T]) Size() int {
	return (*heap[T])(h).size()
}

func (h *MinHeapComparable) ToList() []interface{} {
	return (*heapComparable)(h).tolist()
}

func (h *MinHeap[T]) ToList() []T {
	return (*heap[T])(h).tolist()
}

func (h *MinHeapComparable) Next() interface{} {
	return (*heapComparable)(h).next()
}

func (h *MinHeap[T]) Next() T {
	return (*heap[T])(h).next()
}

func (h *MinHeapComparable) IsEmpty() bool {
	return (*heapComparable)(h).isEmpty()
}

func (h *MinHeap[T]) IsEmpty() bool {
	return (*heap[T])(h).isEmpty()
}

// MaxHeap

func NewMaxHeapComparable() *MaxHeapComparable {
	return (*MaxHeapComparable)(newHeapCapComparable(0))
}

func NewMaxHeap[T constraints.Ordered]() *MaxHeap[T] {
	return (*MaxHeap[T])(newHeapCap[T](0))
}

func NewMaxHeapCapComparable(cap int) *MaxHeapComparable {
	return (*MaxHeapComparable)(newHeapCapComparable(cap))
}

func NewMaxHeapCap[T constraints.Ordered](cap int) *MaxHeap[T] {
	return (*MaxHeap[T])(newHeapCap[T](cap))
}

func (h *MaxHeapComparable) Insert(val interface{}) {
	(*heapComparable)(h).insert(val, true)
}

func (h *MaxHeap[T]) Insert(val T) {
	(*heap[T])(h).insert(val, true)
}

func (h *MaxHeapComparable) Pop() interface{} {
	return (*heapComparable)(h).pop(true)
}

func (h *MaxHeap[T]) Pop() T {
	return (*heap[T])(h).pop(true)
}

func (h *MaxHeapComparable) Size() int {
	return (*heapComparable)(h).size()
}

func (h *MaxHeap[T]) Size() int {
	return (*heap[T])(h).size()
}

func (h *MaxHeapComparable) ToList() []interface{} {
	return (*heapComparable)(h).tolist()
}

func (h *MaxHeap[T]) ToList() []T {
	return (*heap[T])(h).tolist()
}

func (h *MaxHeapComparable) Next() interface{} {
	return (*heapComparable)(h).next()
}

func (h *MaxHeap[T]) Next() T {
	return (*heap[T])(h).next()
}

func (h *MaxHeapComparable) IsEmpty() bool {
	return (*heapComparable)(h).isEmpty()
}

func (h *MaxHeap[T]) IsEmpty() bool {
	return (*heap[T])(h).isEmpty()
}
