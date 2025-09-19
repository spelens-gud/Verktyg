# cahcew 缓存中间件封装

我们在进行开发中经常会遇到这样子的场景：

1. 访问redis缓存
2. 如果redis有缓存 则使用redis缓存字节 反序列化为结构体
3. 如果redis没有缓存 则访问mysql 获取数据
4. 将mysql返回的结果 序列化为字节 并设置redis缓存

我们将以上组件抽象为三部分：

1. 缓存`Store`具有`Get`和`Set`方法
2. 回源闭包 `func() (interface{},error)`
3. 序列化`Marshal`和反序列化`Unmarshal`方法

所以我们可以将以上过程抽象成为下面封装：

```go
package cache

type Store interface {
	Get(key string) ([]byte, error)
	Set(key string, data []byte) error
}

type Cache struct {
	store     Store
	unmarshal func([]byte, interface{}) error
	marshal   func(interface{}) ([]byte, error)
}

func (c Cache) Load(key string, ret interface{}, src func() (interface{}, error)) (err error) {
	// 从缓存获取
	data, err := c.store.Get(key)
	if err == nil {
		// 反序列化
		if err = c.unmarshal(data, &ret); err == nil {
			return
		}
	}
	// 回源
	r, err := src()
	if err != nil {
		return
	}

	// 为保证回源所得结果和从缓存反序列化结果 保持在程序运行时的一致
	// 因此在回源完成后会先将结果进行序列化 然后再反序列化到结构体指针

	// 序列化
	data, err = c.marshal(r)
	if err != nil {
		return
	}

	// 反序列化到指针
	if err = c.unmarshal(data, &ret); err != nil {
		return
	}

	// 设置缓存
	if err = c.store.Set(key, data); err != nil {
		return
	}
	return
}
```

实际代码使用大概如下：

```go

package main

import "strconv"

var cache Cache

// 提供外部调用 带缓存层的封装
func GetUserByUid(uid int) (u User, err error) {
	err = cache.Load("cache:user:uid:"+strconv.Itoa(uid), &u, func() (interface{}, error) {
		return GetUserByUidFromMysql(uid)
	})
	return
}

// 内部实际从sql获取的逻辑
func GetUserByUidFromMysql(uid int) (u User, err error) {
	// ... sql逻辑

	return
}
```

通过以上封装 我们就可以将混杂缓存逻辑的代码拆分得较为干净 同时复用缓存逻辑

适用类似以上场景还有:

1. 获取微信公众号等授权Token
2. 请求其他数据库源
3. ...

以上封装只是一个雏形，实际上在生产使用中，我们还需要考虑以下问题：

0. 统一业务模块缓存前缀
1. 当缓存过期时，如果接口并发量较高，会导致缓存穿透
    1. 在同个key回源请求已发起时，通过全局锁避免并发同key回源，回源后共享结果
2. 回源过程 如果上游发生中断 是否要在后台保持进行回源
    1. 保持回源：当次请求失败，但后台回源设置缓存成功后 后续请求可以使用缓存
    2. 不保持回源：当次请求失败，后续请求仍有可能失败
3. 使用同步更新缓存还是异步更新缓存
    1. 同步更新缓存 缓存设置失败 调用会返回失败
    2. 异步更新缓存 回源成功则能返回结果 但有可能设置缓存失败 通过重试策略尽量保证设置成功
4. 异步更新/删除缓存时 因为重试导致的版本时序问题
    1. 通过更新缓存版本锁控制 保证同key只会由最新版本协程进行更新缓存
5. 提供删除接口
    1. 同步删除 用于保证某些缓存事务一致性场景
    2. 异步双删 能最大限度保证缓存一致性

`CacheW` 提供以下接口封装

其中

```go
package cachew

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
```

其中 `Get` `Set` `Delete` 为对缓存中间件的调用 （`Set` 和 `Delete` 主要添加了进程内同Key版本锁，以防退避重试导致版本时序问题）

`DeleteAsync` 提供异步双删功能 能够在删除key后 在指定间隔后进行再次删除 并通过自动退避重试最大保证删除成功

`Load` 则为 回源自动缓存的封装 与`LoadSyncCache`不同在于

前者回源序列化成功后进行异步更新缓存 并在异步更新失败时自动重试

后者设置缓存为同步进行 在设置缓存失败时会直接返回调用错误 

更多具体 [使用例子](./example_test.go)