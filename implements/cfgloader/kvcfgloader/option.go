package kvcfgloader

import (
	"context"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/spelens-gud/Verktyg/interfaces/istore"
)

type acmKVOption struct {
	keySplit func(string) (group, dataID string)
}

type AcmKVOption func(opt *acmKVOption)

type AcmKVOptions []AcmKVOption

func (opts AcmKVOptions) Init() *acmKVOption {
	o := &acmKVOption{
		keySplit: DefaultSplitConfig,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func WithSplitFunc(f func(string) (group, dataID string)) AcmKVOption {
	return func(opt *acmKVOption) {
		opt.keySplit = f
	}
}

func DefaultSplitConfig(s string) (group, dataID string) {
	spKey := strings.Split(s, "###")
	if len(spKey) != 2 {
		return s, ""
	}
	return spKey[0], spKey[1]
}

func (opts StoreOptions) Init() *storeOption {
	o := &storeOption{
		context:            context.Background(),
		writeFileInterval:  time.Minute * 5,
		clearInterval:      time.Minute * 5,
		defaultExpiration:  time.Minute,
		srcTimeout:         time.Millisecond * 100,
		memStoreExpiration: -1,
		persistenceFile:    "",
		blockNums:          8,
	}
	for _, opt := range opts {
		opt(o)
	}
	if o.store == nil {
		o.store = NewShardCache(o.blockNums, func(i int) istore.FileStore {
			return cache.New(o.defaultExpiration, o.clearInterval)
		})
	}
	return o
}

type storeOption struct {
	context context.Context // 控制内存回收器生命周期的context

	store              istore.FileStore // 储存器
	writeFileInterval  time.Duration    // 持久化间隔
	clearInterval      time.Duration    // 检查内存驻留超时时间
	defaultExpiration  time.Duration    // 默认缓存时间 (缓存时间仅代表多久不从数据源更新数据 但为降级缓存策略 所有数据仍驻留在内存中)
	srcTimeout         time.Duration    // 请求数据源超时时间
	memStoreExpiration time.Duration    // 内存驻留超时
	persistenceFile    string           // 持久化文件名
	blockNums          int              // 分区数
	disableMustGet     bool             // 请求数据源失败且有内存缓存时 不使用降级内存缓存
	lazyUpdate         bool             // 惰性更新 当key过期时不阻塞请求数据源 优先返回旧数据 然后进行异步更新
}

type StoreOption func(*storeOption)

type StoreOptions []StoreOption

func WithPersistenceFile(path string) StoreOption {
	return func(s *storeOption) {
		s.persistenceFile = path
	}
}

func DisableMustGet() StoreOption {
	return func(s *storeOption) {
		s.disableMustGet = true
	}
}

func LazyUpdate() StoreOption {
	return func(s *storeOption) {
		s.lazyUpdate = true
	}
}

func WithContext(ctx context.Context) StoreOption {
	return func(s *storeOption) {
		s.context = ctx
	}
}

func WithMemStoreExpiration(t time.Duration) StoreOption {
	return func(s *storeOption) {
		s.memStoreExpiration = t
	}
}

func WithBlocksNum(shard int) StoreOption {
	return func(s *storeOption) {
		s.blockNums = shard
	}
}

func WithWriteFileInterval(t time.Duration) StoreOption {
	return func(s *storeOption) {
		s.writeFileInterval = t
	}
}

func WithClearInterval(t time.Duration) StoreOption {
	return func(s *storeOption) {
		s.clearInterval = t
	}
}

func WithSrcTimeOut(t time.Duration) StoreOption {
	return func(s *storeOption) {
		s.srcTimeout = t
	}
}

func WithStore(store istore.FileStore) StoreOption {
	return func(s *storeOption) {
		s.store = store
	}
}

func WithDefaultExpiration(t time.Duration) StoreOption {
	return func(s *storeOption) {
		s.defaultExpiration = t
	}
}
