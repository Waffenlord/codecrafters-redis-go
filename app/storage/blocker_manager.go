package storage

import (
	"sync"
)

type BlockerManager struct {
	Clients    map[string][]chan string
	BlockerMux sync.Mutex
}

func NewBlockerManager() *BlockerManager {
	return &BlockerManager{
		Clients:    make(map[string][]chan string),
		BlockerMux: sync.Mutex{},
	}
}

func (b *BlockerManager) BlockedByKey(key string) chan string {
	newChan := make(chan string, 1)
	b.BlockerMux.Lock()
	defer b.BlockerMux.Unlock()
	b.Clients[key] = append(b.Clients[key], newChan)
	return newChan
}

func (b *BlockerManager) NotifyClient(key string, value string) bool {
	b.BlockerMux.Lock()
	defer b.BlockerMux.Unlock()
	chList, ok := b.Clients[key]
	if !ok || len(chList) == 0 {
		return false
	}
	ch := chList[0]
	b.Clients[key] = chList[1:]

	ch <- value
	return true

}
