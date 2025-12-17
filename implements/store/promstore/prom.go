package promstore

import (
	"time"

	"github.com/spelens-gud/Verktyg.git/interfaces/istore"
)

type PromStore struct {
	name    string
	Store   istore.Store
	metrics *storeMetrics
}

func NewPromStore(name string, store istore.Store) *PromStore {
	return &PromStore{name: name, Store: store, metrics: NewStoreMetrics(name)}
}

func (p *PromStore) Get(key string) (value string, err error) {
	t := time.Now()
	defer func() {
		p.metrics.Observe("get", err, time.Since(t))
	}()
	return p.Store.Get(key)
}

func (p *PromStore) Set(key, value string) (err error) {
	t := time.Now()
	defer func() {
		p.metrics.Observe("set", err, time.Since(t))
	}()
	return p.Store.Set(key, value)
}

func (p *PromStore) Add(key, value string) (err error) {
	t := time.Now()
	defer func() {
		p.metrics.Observe("add", err, time.Since(t))
	}()
	return p.Store.Add(key, value)
}

func (p *PromStore) Delete(key string) (err error) {
	t := time.Now()
	defer func() {
		p.metrics.Observe("delete", err, time.Since(t))
	}()
	return p.Store.Delete(key)
}

var _ istore.Store = &PromStore{}
