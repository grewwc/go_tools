package cw

import "github.com/grewwc/go_tools/src/typesw"

const (
	black = iota
	red
)

type treeNode[T any] struct {
	val         T
	left, right *treeNode[T]
	parent      *treeNode[T]
	color       int
}

type RbTree[T any] struct {
	root *treeNode[T]
	cmp  typesw.CompareFunc[T]
	size int
}

func newTreeNode[T any](val T) *treeNode[T] {
	return &treeNode[T]{
		val:    val,
		left:   nil,
		right:  nil,
		parent: nil,
		color:  red,
	}
}

func NewRbTree[T any](cmp typesw.CompareFunc[T]) *RbTree[T] {
	if cmp == nil {
		cmp = typesw.CreateDefaultCmp[T]()
	}
	return &RbTree[T]{
		root: nil,
		cmp:  cmp,
	}
}

func (t *RbTree[T]) Contains(val T) bool {
	curr := t.root

	for curr != nil {
		tmp := t.cmp(val, curr.val)
		if tmp < 0 {
			curr = curr.left
		} else if tmp > 0 {
			curr = curr.right
		} else {
			return true
		}
	}
	return false
}

func (t *RbTree[T]) Search(val T) *treeNode[T] {
	curr := t.root

	for curr != nil {
		tmp := t.cmp(val, curr.val)
		if tmp < 0 {
			curr = curr.left
		} else if tmp > 0 {
			curr = curr.right
		} else {
			return curr
		}
	}
	return nil
}

func (t *RbTree[T]) Insert(val T) {
	newNode := newTreeNode(val)
	curr := t.root
	t.size++
	if curr == nil {
		newNode.color = black
		t.root = newNode
		return
	}
	parent := curr.parent
	for curr != nil {
		parent = curr
		tmp := t.cmp(val, curr.val)
		if tmp < 0 {
			curr = curr.left
		} else {
			curr = curr.right
		}
	}
	newNode.parent = parent
	tmp := t.cmp(val, parent.val)
	if tmp < 0 {
		parent.left = newNode
	} else {
		parent.right = newNode
	}
	t.fixInsert(newNode)
}

func (t *RbTree[T]) fixInsert(node *treeNode[T]) {
	if node.parent == nil || node.parent.color == black {
		return
	}
	gp := node.parent.parent
	if gp == nil {
		return
	}
	parent := node.parent
	var uncle *treeNode[T]
	if parent == gp.left {
		uncle = gp.right
	} else {
		uncle = gp.left
	}
	// case 1: parent & uncle both are "red"
	if uncle != nil && uncle.color == red {
		uncle.color = black
		parent.color = black
		gp.color = red
		t.fixInsert(gp)
		return
	}
	// case 2: "triangle"
	if gp.right == parent && parent.left == node {
		t.rightRotate(parent)
		t.leftRotate(gp)
		node.color = black
		gp.color = red
		return
	}
	if gp.left == parent && parent.right == node {
		t.leftRotate(parent)
		t.rightRotate(gp)
		node.color = black
		gp.color = red
		return
	}
	// case 3: "straight line"
	if gp.right == parent && parent.right == node {
		t.leftRotate(gp)
		parent.color = black
		gp.color = red
		return
	}
	if gp.left == parent && parent.left == node {
		t.rightRotate(gp)
		parent.color = black
		gp.color = red
		return
	}
	// 确保根节点为黑色
	if gp.parent == nil {
		gp.color = black
	}
}

func (t *RbTree[T]) leftRotate(n *treeNode[T]) {
	// 检查边界条件
	if n == nil || n.right == nil {
		return
	}

	// 获取相关节点
	parent := n.parent
	rl := n.right.left
	r := n.right

	// 执行左旋操作
	n.right = rl
	if rl != nil {
		rl.parent = n
	}

	r.left = n
	n.parent = r

	// 更新父节点的指向
	if parent == nil {
		// 如果 n 是根节点，则 r 成为新的根节点
		r.parent = nil
		t.root = r
	} else {
		// 更新父节点的子节点指向
		if parent.left == n {
			parent.left = r
		} else {
			parent.right = r
		}
		// 更新 r 的父节点
		r.parent = parent
	}
}

func (t *RbTree[T]) rightRotate(n *treeNode[T]) {
	// 检查边界条件
	if n == nil || n.left == nil {
		return
	}

	// 获取相关节点
	parent := n.parent
	lr := n.left.right
	l := n.left

	// 执行右旋操作
	l.right = n
	n.parent = l

	n.left = lr
	if lr != nil {
		lr.parent = n
	}

	// 更新父节点的指向
	if parent == nil {
		// 如果 n 是根节点，则 l 成为新的根节点
		l.parent = nil
		t.root = l
	} else {
		// 更新父节点的子节点指向
		if parent.left == n {
			parent.left = l
		} else {
			parent.right = l
		}
		// 更新 l 的父节点
		l.parent = parent
	}
}

func (t *RbTree[T]) Delete(val T) {
	node := t.search(t.root, val)
	if node == nil {
		return // 节点不存在，直接返回
	}
	t.deleteNode(node)
	t.size--
}

func (t *RbTree[T]) search(n *treeNode[T], val T) *treeNode[T] {
	tmp := t.cmp(val, n.val)
	if n == nil || tmp == 0 {
		return n
	}
	if tmp < 0 {
		return t.search(n.left, val)
	}
	return t.search(n.right, val)
}

func (t *RbTree[T]) deleteNode(z *treeNode[T]) {
	var x, y *treeNode[T]
	if z.left == nil || z.right == nil {
		y = z // 如果 z 最多只有一个子节点，则直接删除 z
	} else {
		y = t.minimum(z.right) // 否则找到 z 的后继节点
	}

	if y.left != nil {
		x = y.left
	} else {
		x = y.right
	}

	if x != nil {
		x.parent = y.parent
	}

	if y.parent == nil {
		t.root = x // 如果 y 是根节点，则 x 成为新的根节点
	} else if y == y.parent.left {
		y.parent.left = x
	} else {
		y.parent.right = x
	}

	if y != z {
		z.val = y.val // 替换 z 的值为 y 的值
	}

	if y.color == black {
		t.fixDelete(x, y.parent) // 如果删除的是黑色节点，需要修复红黑树性质
	}
}

func (t *RbTree[T]) minimum(n *treeNode[T]) *treeNode[T] {
	for n.left != nil {
		n = n.left
	}
	return n
}

func (t *RbTree[T]) fixDelete(x, parent *treeNode[T]) {
	for x != t.root && (x == nil || x.color == black) {
		if x == parent.left {
			sibling := parent.right
			if sibling == nil {
				break
			}
			if sibling.color == red { // Case 1: 兄弟节点是红色
				sibling.color = black
				parent.color = red
				t.leftRotate(parent)
				sibling = parent.right
			}
			if sibling != nil {
				if (sibling.left == nil || sibling.left.color == black) &&
					(sibling.right == nil || sibling.right.color == black) {
					// Case 2: 兄弟节点及其子节点都是黑色
					sibling.color = red
					x = parent
					parent = x.parent
				} else {
					if sibling.right == nil || sibling.right.color == black {
						// Case 3: 兄弟节点的左子节点是红色，右子节点是黑色
						sibling.left.color = black
						sibling.color = red
						t.rightRotate(sibling)
						sibling = parent.right
					}
					// Case 4: 兄弟节点的右子节点是红色
					sibling.color = parent.color
					parent.color = black
					if sibling.right != nil {
						sibling.right.color = black
					}
					t.leftRotate(parent)
					x = t.root
				}
			}
		} else {
			sibling := parent.left
			if sibling == nil {
				break
			}
			if sibling.color == red { // Case 1: 兄弟节点是红色
				sibling.color = black
				parent.color = red
				t.rightRotate(parent)
				sibling = parent.left
			}
			if sibling != nil {
				if (sibling.left == nil || sibling.left.color == black) &&
					(sibling.right == nil || sibling.right.color == black) {
					// Case 2: 兄弟节点及其子节点都是黑色
					sibling.color = red
					x = parent
					parent = x.parent
				} else {
					if sibling.left == nil || sibling.left.color == black {
						// Case 3: 兄弟节点的右子节点是红色，左子节点是黑色
						sibling.right.color = black
						sibling.color = red
						t.leftRotate(sibling)
						sibling = parent.left
					}
					// Case 4: 兄弟节点的左子节点是红色
					sibling.color = parent.color
					parent.color = black
					if sibling.left != nil {
						sibling.left.color = black
					}
					t.rightRotate(parent)
					x = t.root
				}
			}
		}
	}
	if x != nil {
		x.color = black
	}
}

// SearchRange search vals in ranges between [lower, upper].
// Both inclusive.
func (t *RbTree[T]) SearchRange(lower, upper T) []T {
	st := NewDeque()
	curr := t.root
	ret := make([]T, 0, 16)
	for !st.Empty() || curr != nil {
		for curr != nil {
			st.PushBack(curr)
			curr = curr.left
		}
		curr = st.PopBack().(*treeNode[T])
		tmp := t.cmp(curr.val, lower)
		if tmp >= 0 {
			if t.cmp(curr.val, upper) <= 0 {
				ret = append(ret, curr.val)
			} else {
				return ret
			}
		}
		curr = curr.right
	}
	return ret
}

func (t *RbTree[T]) Clear() {
	t.size = 0
	t.root = nil
}

func (t *RbTree[T]) Iter() typesw.IterableT[T] {
	f := func() chan T {
		ret := make(chan T)
		go func() {
			defer close(ret)
			st := NewDeque()
			curr := t.root
			for curr != nil || !st.Empty() {
				for curr != nil {
					st.PushBack(curr)
					curr = curr.left
				}
				curr = st.PopBack().(*treeNode[T])
				ret <- curr.val
				curr = curr.right
			}
		}()
		return ret
	}
	return typesw.FuncToIterable(f)
}
