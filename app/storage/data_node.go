package storage

import (
	"errors"
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
	Tree        *RadixNode
	mux         sync.RWMutex
	LastEntryId string
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

func NewStreamType() *StreamType {
	return &StreamType{
		mux: sync.RWMutex{},
	}
}

func (st *StreamType) findEdgePosition(label byte, rn *RadixNode) (int, bool) {
	low := 0
	high := len(rn.Edges) - 1

	for low <= high {
		mid := (low + high) / 2

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

func (st *StreamType) insert(n *RadixNode, key string, value []string) {
	for {
		if len(n.Key) == 0 && len(n.Edges) > 0 {
			commonCharNum := 0
			for _, edge := range n.Edges {
				childNode := edge.child
				commonCharNum = lcp(childNode.Key, key)
				if commonCharNum > 0 {
					n = childNode
					break
				}
			}
			if commonCharNum == 0 {
				newNode := &RadixNode{
					Key:   key,
					IsEnd: true,
					Value: value,
				}
				n.Edges = append(n.Edges, Edge{label: key[0], child: newNode})
			}
		}
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
		} else if common == len(n.Key) && common == len(key) {
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

func (st *StreamType) Add(key string, value []string) (string, error) {
	st.mux.Lock()
	defer st.mux.Unlock()
	if st.Tree == nil {
		validId, err := isStreamAddEntryIdValid(key, "0-0")
		if err != nil {
			return "", err
		}
		st.Tree = &RadixNode{
			Key:   validId,
			IsEnd: true,
			Value: value,
			Edges: nil,
		}
		st.LastEntryId = validId
		return validId, nil
	}

	validId, err := isStreamAddEntryIdValid(key, st.LastEntryId)
	if err != nil {
		return "", err
	}
	st.insert(st.Tree, validId, value)
	st.LastEntryId = validId
	return validId, nil
}

type NodeResult struct {
	Id     string
	Values []string
}

func rangeScan(n *RadixNode, prefix string, start string, end string, result *[]NodeResult) {
	current := prefix + n.Key

	if n.IsEnd {
		if compareStreamIds(current, start, greaterEqual) && compareStreamIds(current, end, lowerEqual) {
			*result = append(*result, NodeResult{Id: current, Values: n.Value})
		}
	}

	for _, edge := range n.Edges {
		rangeScan(edge.child, current, start, end, result)
	}

}

func (st *StreamType) XRange(startId string, endId string) ([]NodeResult, error) {
	st.mux.RLock()
	defer st.mux.RUnlock()
	validStartId, validEndId := validateXRangeIds(startId, endId)

	var results []NodeResult

	if st.Tree == nil {
		return results, errors.New("stream has no entries")
	}
	rangeScan(st.Tree, "", validStartId, validEndId, &results)

	return results, nil
}

func readScan(n *RadixNode, prefix string, start string, result *[]NodeResult) {
	current := prefix + n.Key

	if n.IsEnd {
		if compareStreamIds(current, start, greater) {
			*result = append(*result, NodeResult{Id: current, Values: n.Value})
		}
	}

	for _, edge := range n.Edges {
		readScan(edge.child, current, start, result)
	}
}

type XReadResult struct {
	StreamKey string
	Results   []NodeResult
}

func (st *StreamType) XRead(startId string, streamKey string) (XReadResult, error) {
	st.mux.RLock()
	defer st.mux.RUnlock()

	isValidId := isStreamIdValid(startId)
	if !isValidId {
		return XReadResult{}, errors.New("invalid start id for read command")
	}

	var results []NodeResult
	readScan(st.Tree, "", startId, &results)

	xReadResult := XReadResult{
		StreamKey: streamKey,
		Results:   results,
	}

	return xReadResult, nil	
}