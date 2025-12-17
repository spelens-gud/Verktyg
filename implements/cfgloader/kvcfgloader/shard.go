package kvcfgloader

import (
	"errors"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.uber.org/multierr"

	"github.com/spelens-gud/Verktyg.git/interfaces/istore"
)

type ShardCache struct {
	store []istore.FileStore
}

func defaultHashShard(key string) (block int) {
	ieee := crc32.NewIEEE()
	_, _ = io.WriteString(ieee, key)
	crc := ieee.Sum32()
	return int(crc)
}

func NewShardCache(shard int, factory func(i int) istore.FileStore) *ShardCache {
	shards := make([]istore.FileStore, 0, shard)
	for i := 0; i < shard; i++ {
		shards = append(shards, factory(i))
	}
	return &ShardCache{store: shards}
}

func (s *ShardCache) getBlock(k string) istore.FileStore {
	return s.store[defaultHashShard(k)%len(s.store)]
}

func (s *ShardCache) Set(k string, x interface{}, d time.Duration) {
	s.getBlock(k).Set(k, x, d)
}

func (s *ShardCache) Get(k string) (interface{}, bool) {
	return s.getBlock(k).Get(k)
}

func (s *ShardCache) Delete(key string) {
	s.getBlock(key).Delete(key)
}

func (s *ShardCache) LoadFile(fname string) error {
	var errs []error
	for i, b := range s.store {
		n := fname + "#" + strconv.Itoa(i)
		if _, fileErr := os.Stat(n); fileErr != nil {
			continue
		}
		e := b.LoadFile(n)
		if e != nil {
			errs = append(errs, e)
		}
	}
	return multierr.Combine(errs...)
}

func (s *ShardCache) SaveFile(fname string) error {
	if len(fname) == 0 {
		return errors.New("invalid persistence filename")
	}
	_ = os.MkdirAll(filepath.Dir(fname), 0775)
	var str []string
	for i, b := range s.store {
		e := b.SaveFile(fname + "#" + strconv.Itoa(i))
		if e != nil {
			str = append(str, e.Error())
		}
	}
	if len(str) > 0 {
		return errors.New(strings.Join(str, ":"))
	}
	return nil
}

var _ istore.FileStore = &ShardCache{}
