package storage

import "sync"

type Storage struct {
	data map[string]DataNode
	mux  sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		data: make(map[string]DataNode),
		mux: sync.RWMutex{},
	}
}

func (s *Storage) Set(key string, value DataNode) {
	s.mux.Lock()
	s.data[key] = value
	s.mux.Unlock()
}

func (s *Storage) Get(key string) (DataNode, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	v, ok := s.data[key]
	if !ok {
		return nil, false
	}
	return v, true
}
