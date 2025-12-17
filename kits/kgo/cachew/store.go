package cachew

import (
	"context"
	"crypto/md5"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/spelens-gud/Verktyg/kits/kcontext"
	"github.com/spelens-gud/Verktyg/kits/klog/logger"
)

const logTag = "CACHE_W"

type CacheW interface {
	// 获取缓存Key
	Get(ctx context.Context, key string) ([]byte, error)
	// 直接设置缓存Key
	Set(ctx context.Context, key string, data []byte, exp time.Duration) error
	// 删除缓存
	Delete(ctx context.Context, key string) (err error)
	// 异步双删
	DeleteAsync(ctx context.Context, key string, backoffBase time.Duration)
	// 加载缓存 尝试回源并 异步式设置缓存
	Load(ctx context.Context, key string, exp time.Duration, in interface{}, src func(ctx context.Context) (interface{}, error)) (err error)
	// 加载缓存 尝试回源并 同步式设置缓存 若缓存设置失败 会直接返回调用失败
	LoadSyncCache(ctx context.Context, key string, exp time.Duration, in interface{}, src func(ctx context.Context) (interface{}, error)) (err error)
}

type Store interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, data []byte, exp time.Duration) error
	Delete(ctx context.Context, key string) (err error)
	KeyNotFound(error) bool
}

type cacheW struct {
	option *Option
	store  Store
	sf     *singleflight.Group

	mu       *sync.Mutex
	updating map[string]cacheUpdateVersion
}

type cacheUpdateVersion struct {
	version string
	cancel  func()
}

func NewCacheW(store Store, opts ...func(option *Option)) CacheW {
	opt := options(opts).Init(defaultOption())
	return &cacheW{
		option:   opt,
		store:    store,
		sf:       new(singleflight.Group),
		mu:       new(sync.Mutex),
		updating: make(map[string]cacheUpdateVersion),
	}
}

func (s *cacheW) Delete(ctx context.Context, key string) (err error) {
	key += s.option.KeyPrefix
	return s.delete(ctx, key)
}

func (s *cacheW) delete(ctx context.Context, key string) (err error) {
	if !s.option.DisableUpdateMutex {
		s.mu.Lock()
		// 如果有其他协程在尝试更新缓存 直接删除解锁
		if updating, ok := s.updating[key]; ok {
			updating.cancel()
			delete(s.updating, key)
		}
		s.mu.Unlock()
	}

	return s.store.Delete(ctx, key)
}

func (s *cacheW) DeleteAsync(ctx context.Context, key string, backoffBase time.Duration) {
	key += s.option.KeyPrefix

	go func(ctx context.Context) {
		var (
			lg = logger.FromContext(ctx).WithTag(logTag).WithField("key", key)

			clearErr error
			tryTimes int
		)

		if backoffBase == 0 {
			backoffBase = time.Second
		}

		// 先尝试删除一次
		_ = s.delete(ctx, key)

		// 循环重试删除
		for {
			tryTimes++
			backoffTime := time.Duration(2<<tryTimes) * backoffBase
			time.Sleep(backoffTime)
			if clearErr = s.delete(ctx, key); clearErr == nil {
				return
			}
			lg.Infof("delete cache error: %v", clearErr)
			if tryTimes > s.option.MaxUpdateCacheRetries {
				return
			}
		}
	}(kcontext.Detach(ctx))
}

func (s *cacheW) Get(ctx context.Context, key string) ([]byte, error) {
	key += s.option.KeyPrefix
	return s.store.Get(ctx, key)
}

func (s *cacheW) Set(ctx context.Context, key string, data []byte, exp time.Duration) error {
	key += s.option.KeyPrefix
	return s.updateCache(ctx, key, data, exp, 0, true)
}

func (s *cacheW) loadSourceAndUpdateCache(ctx context.Context, key string, exp time.Duration, in interface{}, src func(ctx context.Context) (interface{}, error), syncUpdate bool) (data []byte, err error) {
	// 回源请求
	ret, err := src(ctx)
	if err != nil {
		return
	}

	// 将回源结果转化为[]byte
	switch ret := ret.(type) {
	case []byte:
		data = ret
	case string:
		data = []byte(ret)
	default:
		// 返回了其他类型 先进行序列化
		if data, err = s.option.MarshalFunc(ret); err != nil {
			logger.FromContext(ctx).Warnf("marshal source result data error: %v", err)
			return nil, err
		}
	}

	// 反序列化 成功反序列化才进行更新缓存
	if err = s.option.UnmarshalFunc(data, in); err != nil {
		logger.FromContext(ctx).Warnf("unmarshal from source result data error: %v", err)
		return
	}

	// 更新缓存
	err = s.updateCache(ctx, key, data, exp, s.option.MaxUpdateCacheRetries, syncUpdate)
	return
}

func (s *cacheW) Load(ctx context.Context, key string, exp time.Duration, in interface{}, src func(ctx context.Context) (interface{}, error)) (err error) {
	return s.load(ctx, key, exp, in, src, false)
}

func (s *cacheW) LoadSyncCache(ctx context.Context, key string, exp time.Duration, in interface{}, src func(ctx context.Context) (interface{}, error)) (err error) {
	return s.load(ctx, key, exp, in, src, true)
}

func asyncLoad(ctx context.Context, f func(ctx context.Context) (interface{}, error)) (ret interface{}, err error) {
	retChan := make(chan interface{})
	waitingDone := ctx.Done()

	go func(ctx context.Context) {
		defer close(retChan)
		r, retErr := f(ctx)
		if retErr != nil {
			r = retErr
		}
		select {
		case <-waitingDone:
		case retChan <- r:
		}
	}(kcontext.Detach(ctx))

	select {
	case <-waitingDone:
		return nil, ctx.Err()
	case ret = <-retChan:
		var ok bool
		if err, ok = ret.(error); ok {
			ret = nil
		}
	}
	return
}

func (s *cacheW) load(ctx context.Context, key string, exp time.Duration, in interface{}, src func(ctx context.Context) (interface{}, error), syncUpdateCache bool) (err error) {
	// 添加全局前缀
	key += s.option.KeyPrefix

	lg := logger.FromContext(ctx).WithTag(logTag).WithField("key", key)

	// 获取缓存数据
	if cacheData, storeErr := s.store.Get(ctx, key); storeErr == nil {
		if storeErr = s.option.UnmarshalFunc(cacheData, in); storeErr == nil {
			return
		} else {
			lg.Warnf("unmarshal store cache error: %v", storeErr)
		}
	} else if !s.store.KeyNotFound(storeErr) {
		lg.Warnf("get cache from store error: %v", storeErr)
	}

	ctx = logger.WithContext(ctx, lg)

	marshaled := false
	// 并发单程回源
	data, err, _ := s.sf.Do(key, func() (r interface{}, err error) {
		marshaled = true
		// 异步加载
		if !syncUpdateCache && s.option.AsyncLoad {
			return asyncLoad(ctx, func(ctx context.Context) (interface{}, error) {
				return s.loadSourceAndUpdateCache(ctx, key, exp, in, src, false)
			})
		}
		return s.loadSourceAndUpdateCache(ctx, key, exp, in, src, syncUpdateCache)
	})

	if err != nil {
		return
	}

	if marshaled {
		return
	}

	// 反序列化
	if err = s.option.UnmarshalFunc(data.([]byte), in); err != nil {
		lg.Warnf("unmarshal data error: %v", err)
		return
	}
	return
}

func (s *cacheW) doUpdateCache(ctx context.Context, key string, data []byte, exp time.Duration, maxRetry int) (err error) {
	if exp <= 0 {
		exp = s.option.DefaultExpiration
	}

	tried := -1
	backoffTick := time.NewTimer(0)
	defer backoffTick.Stop()

	for {
		if err = s.store.Set(ctx, key, data, exp); err == nil {
			return
		}

		tried++

		if tried > maxRetry {
			logger.FromContext(ctx).Warnf("update cache error: %v", err)
			return
		}

		// 退避重试
		retryBackoff := 5 * time.Duration(2<<tried) * time.Millisecond
		logger.FromContext(ctx).Warnf("update cache error: %v,retry after [ %s ]", err, retryBackoff.String())

		backoffTick.Reset(retryBackoff)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-backoffTick.C:
		}
	}
}

func (s *cacheW) updateCache(ctx context.Context, key string, data []byte, exp time.Duration, maxRetry int, sync bool) (err error) {
	if !sync {
		ctx = kcontext.Detach(ctx)
	}

	if s.option.DisableUpdateMutex {
		// 禁用版本并发锁检测
		if sync {
			return s.doUpdateCache(ctx, key, data, exp, maxRetry)
		} else {
			go s.doUpdateCache(ctx, key, data, exp, maxRetry) // nolint
			return nil
		}
	}

	version := fmt.Sprintf("%x", md5.Sum(data))

	s.mu.Lock()

	// 检查是否已经有相同的key在更新缓存
	if updating, ok := s.updating[key]; ok {
		// 已有相同版本的更新协程在进行 直接退出
		if updating.version == version {
			s.mu.Unlock()
			return
		}
		// 版本不一致 取消其他协程动作
		updating.cancel()
	}

	// 创建新的context
	ctx, cf := context.WithCancel(ctx)

	// 覆盖为当前版本
	s.updating[key] = cacheUpdateVersion{
		version: version,
		cancel:  cf,
	}
	s.mu.Unlock()

	// 处理完成 解锁
	cleanUp := func() {
		cf()
		s.mu.Lock()
		// 锁没有被其他协程摘掉 且还是自己的版本则 解锁
		if u, exist := s.updating[key]; exist && u.version == version {
			delete(s.updating, key)
		}
		s.mu.Unlock()
	}

	// 真正的更新缓存逻辑
	if sync {
		defer cleanUp()
		return s.doUpdateCache(ctx, key, data, exp, maxRetry)
	} else {
		go func() {
			defer cleanUp()
			_ = s.doUpdateCache(ctx, key, data, exp, maxRetry)
		}()
		return nil
	}
}
