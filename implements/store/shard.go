package store

import (
	"hash/crc32"
	"io"

	"github.com/spelens-gud/Verktyg.git/interfaces/istore"
)

type ShardStore struct {
	hashFunc    func(key string) (block int)
	connections []istore.Store
	l           int
}

type ShardStoreOpt func(*ShardStore)

func defaultHashShard(key string) (block int) {
	ieee := crc32.NewIEEE()
	_, _ = io.WriteString(ieee, key)
	crc := ieee.Sum32()
	return int(crc)
}

func WithHashShard(f func(key string) (block int)) ShardStoreOpt {
	return func(store *ShardStore) {
		store.hashFunc = f
	}
}

func NewShardStoreFromAddr(addrs []string, storeFactory func(addr string) (istore.Store, error), opt ...ShardStoreOpt) (istore.Store, error) {
	store := &ShardStore{
		connections: make([]istore.Store, 0, len(addrs)),
		l:           len(addrs),
		hashFunc:    defaultHashShard,
	}

	for _, o := range opt {
		o(store)
	}

	for _, addr := range addrs {
		st, err := storeFactory(addr)
		if err != nil {
			return nil, err
		}
		store.connections = append(store.connections, st)
	}
	return store, nil
}

func (b *ShardStore) getBlock(key string) istore.Store {
	return b.connections[b.hashFunc(key)%b.l]
}

func (b *ShardStore) Get(key string) (value string, err error) {
	return b.getBlock(key).Get(key)
}

func (b *ShardStore) Set(key, value string) (err error) {
	return b.getBlock(key).Set(key, value)
}

func (b *ShardStore) Delete(key string) (err error) {
	return b.getBlock(key).Delete(key)
}

func (b *ShardStore) Add(key, value string) (err error) {
	return b.getBlock(key).Add(key, value)
}

var _ istore.Store = &ShardStore{}
