package cw

import "github.com/grewwc/go_tools/src/algow"

type BitMap struct {
	data []uint32
}

func NewBitMap(size int) *BitMap {
	return &BitMap{
		data: make([]uint32, size/32+1),
	}
}

func (m *BitMap) SetBit(pos int) bool {
	idx, mod := m.divmod(pos)
	if idx >= len(m.data) || pos < 0 {
		return false
	}
	m.data[idx] |= 1 << mod
	return false
}

func (m *BitMap) ClearBit(pos int) bool {
	idx, mod := m.divmod(pos)
	if idx >= len(m.data) || pos < 0 {
		return false
	}
	m.data[idx] &= ^(1 << mod)
	return true
}

func (m *BitMap) FlipBit(pos int) bool {
	idx, mod := m.divmod(pos)
	if idx >= len(m.data) || pos < 0 {
		return false
	}
	m.data[idx] ^= (1 << mod)
	return true
}

func (m *BitMap) Clear() {
	algow.Fill(m.data, 0)
}

// TestBit returns true if `pos` is set
func (m *BitMap) TestBit(pos int) bool {
	idx, mod := m.divmod(pos)
	if idx >= len(m.data) || pos < 0 {
		return false
	}
	return m.data[idx]&(1<<mod) != 0
}

func (m *BitMap) divmod(pos int) (int, int) {
	return pos / 32, pos % 32
}
