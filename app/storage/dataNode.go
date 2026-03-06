package storage

import "time"

type DataNode interface {
	isDataNode()
}

type StringType struct {
	Value     string
	CreatedAt time.Time
	ExpMil    int64
}

func (st *StringType) isDataNode() {}

type ListType struct {
	Head *ListNode
	Tail *ListNode
	Len  int
}

type ListNode struct {
	next     *ListNode
	previous *ListNode
	value    string
}

func (lt *ListType) isDataNode() {}

func NewList(value string) *ListType {
	firstNode := ListNode{
		value: value,
	}
	newList := ListType{
		Head: &firstNode,
		Tail: &firstNode,
		Len:  1,
	}
	return &newList
}

func (lt *ListType) AppendR(value string) int {
	currentLast := lt.Tail
	newNode := ListNode{
		value:    value,
		previous: currentLast,
	}
	currentLast.next = &newNode
	lt.Len++
	lt.Tail = &newNode
	return lt.Len
}
