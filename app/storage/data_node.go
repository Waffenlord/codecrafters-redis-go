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

func NewList() *ListType {
	newList := &ListType{
		Len: 0,
		Mux: sync.RWMutex{},
	}
	return newList
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

func (lt *ListType) Lpop(n int) []string {
	lt.Mux.Lock()
	defer lt.Mux.Unlock()
	limit := min(n, lt.Len)
	result := []string{}
	for range limit {
		if lt.Head == nil {
			return result
		}
		node := lt.Head
		lt.Head = lt.Head.next
		if lt.Head != nil {
			lt.Head.previous = nil
		}
		lt.Len--
		result = append(result, node.value)
	}
	return result
}

type StreamType struct {
	Tree *RadixNode
	mux  sync.RWMutex
}

type RadixNode struct {
	Key   string
	Edges []Edge
	IsEnd bool
	Value []string
}

type Edge struct {
	label byte
	child *RadixNode
}

func (st *StreamType) isDataNode() {}

func NewStreamType(n *RadixNode) *StreamType {
	return &StreamType{
		Tree: n,
		mux:  sync.RWMutex{},
	}
}

func (st *StreamType) findEdgePosition(label byte, rn *RadixNode) (int, bool) {
	low := 0
	high := len(rn.Edges) - 1

	for low < high {
		mid := low + high/2
		if rn.Edges[mid].label < label {
			low = mid + 1
		} else {
			high = mid
		}

	}

	if low < len(rn.Edges) && rn.Edges[low].label == label {
		return low, true
	}

	return low, false
}

func (st *StreamType) addEdge(e Edge, rn *RadixNode) {
	idx, found := st.findEdgePosition(e.label, rn)
	if found {
		rn.Edges[idx] = e
		return
	}

	rn.Edges = append(rn.Edges, Edge{})
	copy(rn.Edges[idx+1:], rn.Edges[idx:])
	rn.Edges[idx] = e
}

func (st *StreamType) Insert(n *RadixNode, key string, value []string) {
	st.mux.Lock()
	defer st.mux.Unlock()
	for {
		common := lcp(n.Key, key)

		if common < len(n.Key) {
			child := &RadixNode{
				Key:   n.Key[common:],
				Edges: n.Edges,
				IsEnd: n.IsEnd,
				Value: n.Value,
			}

			n.Key = n.Key[:common]
			n.Edges = nil
			n.IsEnd = false
			n.Value = nil

			st.addEdge(Edge{
				label: child.Key[0],
				child: child,
			}, n)
		}

		if common == len(n.Key) {
			n.IsEnd = true
			n.Value = value
			return
		}

		key = key[common:]

		label := key[0]
		idx, found := st.findEdgePosition(label, n)
		if !found {
			newNode := &RadixNode{
				Key:   key,
				IsEnd: true,
				Value: value,
			}
			st.addEdge(Edge{label: label, child: newNode}, n)
			return
		}

		n = n.Edges[idx].child
	}
}
