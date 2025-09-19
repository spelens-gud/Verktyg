# 基础组件实现

此包内组件均实现了[interfaces](../interfaces)内的某个接口 并且实例化后以接口形式暴露

- [cfgloader](./cfgloader) 提供多种配置加载器的实现 包括`acm` `nacos` `static` 实现了`iconfig.ConfigLoader`
- [gormctx](./gormctx) 支持`context.Context`的魔改GORM 包含链路追踪和聚合监控功能 
- [httpreq](./httpreq) http调用客户端库 优化了使用时url拼接和对返回值处理的API 支持自动重试 包含链路追踪和聚合监控功能
- [logrusentry](./logrusentry) 基于`logrus`实现的`ilog.Logger`
- [otrace](./otrace) 基于`opentracing`实现的链路追踪封装 实现`itrace.Tracer`
- [promdb](./promdb) 基于`prometheus`实现的数据库监控封装 包含两部分 用量监控 和 连接池监控
- [promgateway](./promgateway) 基于`prometheus pushgateway`的上报封装
- [promhttp](./promhttp) 基于`prometheus` 对http服务器和调用的监控封装 
- [promrpc](./promrpc) 基于`prometheus` 对rpc服务器和调用的监控封装 
- [redisx](./redisx) 基于`go-redis/redis/v8`的`redis`调用封装 包含链路追踪和聚合监控功能 
- [skafka](./skafka) 基于`sarama`的`kafka`调用封装 
- [skytrace](./skytrace) 基于`go2sky`的`skywalking`链路追踪封装 实现了 `itrace.Tracer` 
- [store](./store) 提供了多种`KV Store`的实现 包含`memcached`和本地内存储存 实现了`istore.Store` 
- [worker](./worker) 提供了对`crontab`定时任务队列的封装 实现了 `iwork.Worker`
- [yapi](./yapi) 提供了对`yapi`调用API的封装
- [zaplog](./zaplog) 基于`zap`实现的`ilog.Logger`