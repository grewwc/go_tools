package leetcode

import "fmt"

type ListNode struct {
	Val  int
	Next *ListNode
}

func PrintList(head *ListNode) {
	for head != nil {
		fmt.Println(head.Val)
		head = head.Next
	}
}

func NewList(vals ...int) *ListNode {
	res := ListNode{0, nil}
	cur := &res
	for _, val := range vals {
		cur.Next = &ListNode{val, nil}
		cur = cur.Next
	}
	return res.Next
}
