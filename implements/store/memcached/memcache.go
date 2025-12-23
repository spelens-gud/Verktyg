package memcached

import (
	"errors"
	"fmt"
	"time"

	"github.com/bradfitz/gomemcache/memcache"

	"github.com/spelens-gud/Verktyg/implements/store/promstore"
	"github.com/spelens-gud/Verktyg/interfaces/istore"
	"github.com/spelens-gud/Verktyg/kits/kdb"
)

type Store struct {
	Client *memcache.Client
}

var (
	defaultTimeout      = time.Second * 1
	defaultMaxIdleConns = 10
)

func SetOption(timeout time.Duration, maxIdelConns int) {
	defaultTimeout = timeout
	defaultMaxIdleConns = maxIdelConns
}

func New(servers ...string) (istore.Store, error) {
	if len(servers) == 0 {
		return nil, errors.New("invalid servers")
	}

	cli := memcache.New(servers...)
	cli.Timeout = defaultTimeout
	cli.MaxIdleConns = defaultMaxIdleConns

	if err := kdb.PingDB(cli.Ping,
		"memcached", fmt.Sprint(servers)); err != nil {
		return nil, err
	}

	store := &Store{Client: cli}
	return promstore.NewPromStore("memcached:"+servers[0], store), nil
}

func (m *Store) Add(key, value string) (err error) {
	return m.Client.Add(&memcache.Item{Key: key, Value: []byte(value)})
}

func (m *Store) Get(key string) (value string, err error) {
	item, err := m.Client.Get(key)
	if err != nil {
		return "", err
	}
	return string(item.Value), nil
}

func (m *Store) Delete(key string) (err error) {
	return m.Client.Delete(key)
}

func (m *Store) Set(key, value string) (err error) {
	return m.Client.Set(&memcache.Item{Key: key, Value: []byte(value)})
}

func (m *Store) SetEX(key, value string, expireSeconds int32) (err error) {
	return m.Client.Set(&memcache.Item{
		Key:        key,
		Value:      []byte(value),
		Expiration: expireSeconds,
	})
}

var _ istore.Store = &Store{}
