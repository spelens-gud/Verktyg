package kvcfgloader

import (
	"context"
	"encoding/gob"
	"os"
	"path/filepath"
	"time"

	"github.com/spelens-gud/Verktyg.git/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg.git/interfaces/ierror"
	"github.com/spelens-gud/Verktyg.git/interfaces/istore"
	"github.com/spelens-gud/Verktyg.git/kits/kerror/errorx"
	"github.com/spelens-gud/Verktyg.git/kits/klog/logger"
)

type StoreConfigLoader struct {
	src         iconfig.KVConfigSourceLoader
	cache       istore.FileStore
	opt         *storeOption `structgraph:"-"`
	updateQueue chan string
}

type item struct {
	Value    string
	NotFound bool
	Exp      int64
}

var errNotFound = errorx.ErrNotFound("key not found")

func IsKeyNotFound(err error) bool {
	return err == errNotFound
}

func (i item) IsExp() bool {
	return i.Exp > 0 && time.Now().UnixNano() > i.Exp
}

func init() {
	gob.Register(item{})
}

func NewCacheConfigLoader(src iconfig.KVConfigSourceLoader, opts ...StoreOption) *StoreConfigLoader {
	opt := StoreOptions(opts).Init()

	loader := &StoreConfigLoader{
		src:   src,
		cache: opt.store,
		opt:   opt,
	}

	if len(opt.persistenceFile) > 0 {
		_ = os.MkdirAll(filepath.Dir(opt.persistenceFile), 0775)
		err := loader.cache.LoadFile(opt.persistenceFile)
		if err != nil {
			logger.FromBackground().Errorf("load store file error: %v", err)
		}

	}

	if loader.opt.lazyUpdate {
		loader.updateQueue = make(chan string, 2<<16)
		for i := 0; i < 5; i++ {
			go loader.startLazyUpdateDaemon()
		}
	}

	if loader.opt.writeFileInterval > 0 && len(loader.opt.persistenceFile) > 0 {
		go loader.startWriteDaemon()
	}
	return loader
}

// 定期持久化文件夹
func (c *StoreConfigLoader) startLazyUpdateDaemon() {
	ctx := context.Background()
	for key := range c.updateQueue {
		if _, err := c.updateKey(ctx, key); err != nil && !IsKeyNotFound(err) {
			logger.FromContext(ctx).Errorf("update cache error: %v", err)
		}
	}
}

func (c *StoreConfigLoader) startWriteDaemon() {
	ticker := time.NewTicker(c.opt.writeFileInterval)
	defer ticker.Stop()
	for {
		select {
		case <-c.opt.context.Done():
			return
		case <-ticker.C:
			err := c.cache.SaveFile(c.opt.persistenceFile)
			if err != nil {
				logger.FromBackground().Errorf("save store file error: %v", err)
			}
		}
	}
}

func (c *StoreConfigLoader) GetWithContext(ctx context.Context, key string) (value string, err error) {
	var cacheValue string
	ret, hasCache := c.cache.Get(key)
	if hasCache {
		// 有缓存
		if i, ok := ret.(item); ok {
			// 如果是未命中的缓存 返回错误
			cacheValue = i.Value
			if i.NotFound {
				err = errNotFound
			}
			// 未过期 返回Key
			if !i.IsExp() {
				value = cacheValue
				return
			} else if c.opt.lazyUpdate {
				value = cacheValue

				// 尝试入队列
				select {
				case c.updateQueue <- key:
				case <-ctx.Done():
				default:
				}

				return
			}
		}
	}

	// 更新key
	if value, err = c.updateKey(ctx, key); IsKeyNotFound(err) {
		return
	}

	// 默认 读取错误 如果内存有数据 则不返回错误 返回内存缓存
	if err != nil && !c.opt.disableMustGet && hasCache {
		return cacheValue, nil
	}
	return
}

func (c *StoreConfigLoader) updateKey(ctx context.Context, key string) (value string, err error) {
	if c.opt.srcTimeout > 0 {
		var cf func()
		ctx, cf = context.WithTimeout(ctx, c.opt.srcTimeout)
		defer cf()
	}

	// 读取源
	if value, err = c.src.GetWithContext(ctx, key); err == nil {
		// 更新成功 更新缓存
		c.cache.Set(key, item{
			Value: value,
			Exp:   time.Now().Add(c.opt.defaultExpiration).UnixNano(),
		}, c.opt.memStoreExpiration)
		return
	}

	if codeErr, ok := err.(ierror.CodeError); ok {
		// 如果返回的是404 设置 notFound
		switch codeErr.Code() {
		case ierror.NotFound:
			c.cache.Set(key, item{
				NotFound: true,
				Exp:      time.Now().Add(c.opt.defaultExpiration / 2).UnixNano(),
			}, c.opt.memStoreExpiration)
			err = errNotFound
			return
		}
	}
	logger.FromContext(ctx).Errorf("get config %s from src error: %v", key, err)
	return
}

func (c *StoreConfigLoader) SetWithContext(ctx context.Context, key string, value string, exp time.Duration) (err error) {
	if c.opt.srcTimeout > 0 {
		var cf func()
		ctx, cf = context.WithTimeout(ctx, c.opt.srcTimeout)
		defer cf()
	}

	err = c.src.SetWithContext(ctx, key, value)
	if err != nil {
		return
	}
	if exp == 0 {
		exp = c.opt.defaultExpiration
	}

	i := item{Value: value}

	if exp > 0 {
		i.Exp = time.Now().Add(exp).UnixNano()
	}

	c.cache.Set(key, i, c.opt.memStoreExpiration)
	return
}

func (c *StoreConfigLoader) Get(key string) (value string, err error) {
	return c.GetWithContext(context.Background(), key)
}

func (c *StoreConfigLoader) Set(key string, value string, exp time.Duration) (err error) {
	return c.SetWithContext(context.Background(), key, value, exp)
}

var _ iconfig.KVConfigLoader = &StoreConfigLoader{}
