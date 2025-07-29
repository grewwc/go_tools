package cw

import (
	"strings"
)

type HuffmanEncoder struct {
	encodingTable *Map[byte, string]

	lastByteLen int
}

type _huffman_node struct {
	b           _byte
	cnt         int
	left, right *_huffman_node
}

type _byte struct {
	data byte
	len  int
}

func NewHuffmanEncoder() *HuffmanEncoder {
	return &HuffmanEncoder{}
}

// build byte table
func buildTable(root *_huffman_node) *Map[byte, string] {
	table := NewMap[byte, string]()
	stack := NewStack[*Tuple]()
	curr := NewTuple(root, "")
	for !stack.Empty() || curr.Get(0).(*_huffman_node) != nil {
		for curr.Get(0).(*_huffman_node) != nil {
			stack.Push(curr)
			t := NewTuple(curr.Get(0).(*_huffman_node).left, curr.Get(1).(string)+"0")
			curr = t
		}
		curr = stack.Pop()
		b := curr.Get(0).(*_huffman_node).b
		path := curr.Get(1).(string)
		if b.len > 0 {
			table.Put(b.data, path)
		}
		curr = NewTuple(curr.Get(0).(*_huffman_node).right, path+"1")
	}
	return table
}

func (encoder *HuffmanEncoder) strToBytes(str string) []byte {
	encodeSingleByte := func(str string) byte {
		var res byte
		for i := 0; i < 8 && i < len(str); i++ {
			if str[i] == '1' {
				res |= (1 << (7 - i))
			}
		}
		return res
	}
	res := make([]byte, 0, len(str)/8)
	index := 0
	for index < len(str)-8 {
		b := encodeSingleByte(str[index : index+8])
		res = append(res, b)
		index += 8
	}
	encoder.lastByteLen = len(str) - index
	res = append(res, encodeSingleByte(str[index:]))
	return res
}

func (encoder *HuffmanEncoder) bytesToStr(data []byte) string {
	decodeSingleByte := func(b byte) string {
		var res strings.Builder
		res.Grow(8)
		for i := 0; i < 8; i++ {
			if b&(1<<(7-i)) != 0 {
				res.WriteByte('1')
			} else {
				res.WriteByte('0')
			}
		}
		return res.String()
	}
	var res strings.Builder
	res.Grow(len(data) * 8)
	for _, b := range data {
		res.WriteString(decodeSingleByte(byte(b)))
	}
	return res.String()[:res.Len()-8+encoder.lastByteLen]
}

func (encoder *HuffmanEncoder) Encode(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}
	counter := make(map[byte]int)
	for _, d := range data {
		counter[d]++
	}
	h := NewHeap(func(n1, n2 *_huffman_node) int {
		return n1.cnt - n2.cnt
	})
	for b, cnt := range counter {
		n := &_huffman_node{
			b:   _byte{data: b, len: 8},
			cnt: cnt,
		}
		h.Insert(n)
	}
	for h.Size() > 1 {
		n1 := h.Pop()
		n2 := h.Pop()
		newNode := &_huffman_node{
			cnt:   n1.cnt + n2.cnt,
			left:  n1,
			right: n2,
			b:     _byte{len: 0},
		}
		h.Insert(newNode)
	}

	m := buildTable(h.Top())
	encoder.encodingTable = m
	res := strings.Builder{}
	for _, b := range data {
		res.WriteString(m.Get(b))
	}
	// fmt.Println("original", res.String()[:20])
	return encoder.strToBytes(res.String())
}

func (encoder *HuffmanEncoder) Decode(encoded []byte) []byte {
	// build reverse table
	m := encoder.encodingTable.ReverseKV()
	curr := strings.Builder{}

	var res []byte
	encodedString := encoder.bytesToStr(encoded)
	// fmt.Println("after", encodedString[:20])
	for i := 0; i < len(encodedString); i++ {
		curr.WriteByte(encodedString[i])
		if m.Contains(curr.String()) {
			res = append(res, m.Get(curr.String()))
			curr.Reset()
		}
	}
	if curr.Len() != 0 {
		return nil
	}
	return res
}
