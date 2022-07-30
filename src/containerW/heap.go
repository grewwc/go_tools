package containerW

import (
	"github.com/grewwc/go_tools/src/containerW/typesW"
)

type heap struct {
	data []interface{}
}

type Heap interface {
	Insert(interface{})
	Pop() interface{}
	Size() int
	IsEmpty() bool
	ToList() []interface{}
	Next() interface{}
}

type MinHeap heap

type MaxHeap heap

func newHeapCap(cap int) *heap {
	return &heap{data: make([]interface{}, 1, cap+1)}
}

func (h *heap) size() int {
	return len(h.data) - 1
}

func (h *heap) insert(val interface{}, reverse bool) {
	h.data = append(h.data, val)
	swim(h.data, len(h.data)-1, reverse)
}

func (h *heap) next() interface{} {
	if len(h.data) == 1 {
		return nil
	}
	return h.data[1]
}

func (h *heap) tolist() []interface{} {
	res := make([]interface{}, h.size())
	for i, val := range h.data[1:] {
		res[i] = val
	}
	return res
}

func (h *heap) pop(reverse bool) interface{} {
	if h.size() <= 0 {
		return nil
	}
	res := h.next()
	// swap
	h.data[1], h.data[len(h.data)-1] = h.data[len(h.data)-1], h.data[1]
	h.data = h.data[:len(h.data)-1]
	sink(h.data, 1, reverse)
	return res
}

func (h *heap) isEmpty() bool {
	return len(h.data) == 1
}

func swim(arr []interface{}, idx int, reverse bool) {
	if idx <= 1 {
		return
	}
	cmp := arr[idx/2].(typesW.Comparable).Compare(arr[idx])
	if reverse {
		cmp *= -1
	}
	if cmp > 0 {
		arr[idx/2], arr[idx] = arr[idx], arr[idx/2]
		swim(arr, idx/2, reverse)
		sink(arr, idx, reverse)
	}
}

func sink(arr []interface{}, idx int, reverse bool) {
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
	}
	sink(arr, minChild, reverse)
}

// public methods
// NewMinHeap

func NewMinHeap() *MinHeap {
	return (*MinHeap)(newHeapCap(0))
}

func NewMinHeapCap(cap int) *MinHeap {
	return (*MinHeap)(newHeapCap(cap))
}

func (h *MinHeap) Insert(val interface{}) {
	(*heap)(h).insert(val, false)
}

func (h *MinHeap) Pop() interface{} {
	return (*heap)(h).pop(false)
}

func (h *MinHeap) Size() int {
	return (*heap)(h).size()
}

func (h *MinHeap) ToList() []interface{} {
	return (*heap)(h).tolist()
}

func (h *MinHeap) Next() interface{} {
	return (*heap)(h).next()
}

func (h *MinHeap) IsEmpty() bool {
	return (*heap)(h).isEmpty()
}

// MaxHeap

func NewMaxHeap() *MaxHeap {
	return (*MaxHeap)(newHeapCap(0))
}

func NewMaxHeapCap(cap int) *MaxHeap {
	return (*MaxHeap)(newHeapCap(cap))
}

func (h *MaxHeap) Insert(val interface{}) {
	(*heap)(h).insert(val, true)
}

func (h *MaxHeap) Pop() interface{} {
	return (*heap)(h).pop(true)
}

func (h *MaxHeap) Size() int {
	return (*heap)(h).size()
}

func (h *MaxHeap) ToList() []interface{} {
	return (*heap)(h).tolist()
}

func (h *MaxHeap) Next() interface{} {
	return (*heap)(h).next()
}

func (h *MaxHeap) IsEmpty() bool {
	return (*heap)(h).isEmpty()
}
