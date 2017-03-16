package test

import (
	"reflect"
	"sync"
)

type Spy struct {
	sync.Mutex
	Calls map[string][][]interface{}
}

func (s *Spy) Call(funcName string, args []interface{}) {
	s.Lock()
	defer s.Unlock()
	if s.Calls == nil {
		s.Calls = map[string][][]interface{}{}
	}
	if calls, ok := s.Calls[funcName]; ok {
		s.Calls[funcName] = append(calls, args)
	} else {
		s.Calls[funcName] = [][]interface{}{args}
	}
}

func (s *Spy) Clear() {
	s.Lock()
	defer s.Unlock()
	s.Calls = map[string][][]interface{}{}
}

func (s *Spy) WasEverCalled() bool {
	return len(s.Calls) > 0
}

func (s *Spy) WasCalled(funcName string) bool {
	return len(s.Calls[funcName]) > 0
}

func (s *Spy) WasCalledWith(funcName string, funcArgs []interface{}) bool {
	funcCalls, ok := s.Calls[funcName]
	if !ok {
		return false
	}

	for _, args := range funcCalls {
		if reflect.DeepEqual(args, funcArgs) {
			return true
		}
	}

	return false
}
