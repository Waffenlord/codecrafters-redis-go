package storage

import (
	"errors"
	"strconv"
	"sync"
	"time"
)

type Storage struct {
	data    map[string]DataNode
	mux     sync.RWMutex
	Blocker *BlockerManager
}

func NewStorage() *Storage {
	return &Storage{
		data:    make(map[string]DataNode),
		mux:     sync.RWMutex{},
		Blocker: NewBlockerManager(),
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

func (s *Storage) Delete(key string) {
	s.mux.Lock()
	delete(s.data, key)
	s.mux.Unlock()
}

func (s *Storage) Incr(key string) (int, error) {
    s.mux.Lock()
    defer s.mux.Unlock()

    v, found := s.data[key]

    if !found {
        data := &StringType{
            Value:     "1",
            CreatedAt: time.Now(),
        }
        s.data[key] = data
        return 1, nil
    }

    data, ok := v.(*StringType)
    if !ok {
        return 0, errors.New("value stored should be a string type")
    }

    parsed, err := strconv.Atoi(data.Value)
    if err != nil {
        return 0, errors.New("value is not an integer or out of range")
    }

    parsed++
    data.Value = strconv.Itoa(parsed)

    return parsed, nil
}