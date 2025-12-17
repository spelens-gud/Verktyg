package memory

import (
	"sync"

	"github.com/spelens-gud/Verktyg/interfaces/istore"
)

func init() {
	istore.RegisteredStore["memory"] = New
}

type Store struct {
	m map[string]string
	l sync.Mutex
}

func New(_ ...string) (istore.Store, error) {
	return &Store{
		m: map[string]string{},
		l: sync.Mutex{},
	}, nil
}

func (s *Store) Get(key string) (value string, err error) {
	s.l.Lock()
	value = s.m[key]
	s.l.Unlock()
	return
}

func (s *Store) Set(key, value string) (err error) {
	s.l.Lock()
	s.m[key] = value
	s.l.Unlock()
	return
}

func (s *Store) Delete(key string) (err error) {
	s.l.Lock()
	delete(s.m, key)
	s.l.Unlock()
	return
}

func (s *Store) Add(key, value string) (err error) {
	s.l.Lock()
	if _, ok := s.m[key]; !ok {
		s.m[key] = value
	}
	s.l.Unlock()
	return
}

var _ istore.Store = &Store{}
