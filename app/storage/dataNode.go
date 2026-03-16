package storage

import (
	"sync"
	"time"
)

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
	Mux  sync.RWMutex
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
		Mux:  sync.RWMutex{},
	}
	return &newList
}

func (lt *ListType) AppendR(value string) int {
	lt.Mux.Lock()
	defer lt.Mux.Unlock()
	newNode := &ListNode{
		value:    value,
		previous: lt.Tail,
	}

	if lt.Tail == nil {
		lt.Head = newNode
		lt.Tail = newNode
	} else {
		lt.Tail.next = newNode
		lt.Tail = newNode
	}

	lt.Len++
	return lt.Len
}

func (lt *ListType) AppendL(value string) int {
	lt.Mux.Lock()
	defer lt.Mux.Unlock()
	newNode := &ListNode{
		value: value,
		next:  lt.Head,
	}

	if lt.Head == nil {
		lt.Head = newNode
		lt.Tail = newNode
	} else {
		lt.Head.previous = newNode
		lt.Head = newNode
	}
	lt.Len++
	return lt.Len
}

func (lt *ListType) LRange(start int, end int) []string {
	lt.Mux.RLock()
	defer lt.Mux.RUnlock()
	currentNode := lt.Head
	if currentNode == nil {
		return make([]string, 0)
	}
	result := []string{}
	for i := 0; i <= end; i++ {
		if i >= start {
			result = append(result, currentNode.value)
		}
		currentNode = currentNode.next
	}
	return result
}
