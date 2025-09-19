# golang基础组件库


## [implements](./implements) 为组件实现

- [cfgloader](./implements/cfgloader) 提供多种配置加载器的实现 包括`acm` `nacos` `static` 实现了`iconfig.ConfigLoader`
- [gormctx](./implements/gormctx) 支持`context.Context`的魔改GORM 包含链路追踪和聚合监控功能 
- [httpreq](./implements/httpreq) http调用客户端库 优化了使用时url拼接和对返回值处理的API 支持自动重试 包含链路追踪和聚合监控功能
- [logrusentry](./implements/logrusentry) 基于`logrus`实现的`ilog.Logger`
- [otrace](./implements/otrace) 基于`opentracing`实现的链路追踪封装 实现`itrace.Tracer`
- [promdb](./implements/promdb) 基于`prometheus`实现的数据库监控封装 包含两部分 用量监控 和 连接池监控
- [promgateway](./implements/promgateway) 基于`prometheus pushgateway`的上报封装
- [promhttp](./implements/promhttp) 基于`prometheus` 对http服务器和调用的监控封装 
- [promrpc](./implements/promrpc) 基于`prometheus` 对rpc服务器和调用的监控封装 
- [redisx](./implements/redisx) 基于`go-redis/redis/v8`的`redis`调用封装 包含链路追踪和聚合监控功能 
- [skafka](./implements/skafka) 基于`sarama`的`kafka`调用封装 
- [skytrace](./implements/skytrace) 基于`go2sky`的`skywalking`链路追踪封装 实现了 `itrace.Tracer` 
- [store](./implements/store) 提供了多种`KV Store`的实现 包含`memcached`和本地内存储存 实现了`istore.Store` 
- [worker](./implements/worker) 提供了对`crontab`定时任务队列的封装 实现了 `iwork.Worker`
- [yapi](./implements/yapi) 提供了对`yapi`调用API的封装
- [zaplog](./implements/zaplog) 基于`zap`实现的`ilog.Logger`
 
## [interfaces](./interfaces) 为组件抽象接口

- [iconfig](./interfaces/iconfig) 定义了配置加载层的接口 包括如何以什么参数获取配置 以及部分应用功能环境变量 
- [idoc](./interfaces/idoc) 参照`Openapi`定义了接口文档规范 包括 `Json schema`的相关定义 后续应被官方定义替代
- [ierror](./interfaces/ierror) 定义了错误类型以及错误码对http服务状态码的映射 参考`grpc`定义
- [ihttp](./interfaces/ihttp) 定义了http调用的接口 以及重定义了系统默认的http调用client
- [ikafka](./interfaces/ikafka) 定义了基于`sarama`的`kafka`调用接口和对象构造关系
- [ilog](./interfaces/ilog) 定义了日志模块的规范 包括部分Tag以及日志级别等
- [imetrics](./interfaces/imetrics) 定义了聚合监控的调用规范 以及部分命名空间 系统划分
- [iredis](./interfaces/iredis) 定义了`redis`调用的规范 包括构造时需要传入的参数 以及如何提供依赖
- [isql](./interfaces/isql) 定义了`sql`调用的规范 包括构造时需要传入的参数 以及如何提供依赖
- [istore](./interfaces/istore) 定义了一个`KVStore`的抽象层 以及构造参数
- [itest](./interfaces/itest) 定义了http接口测试的逻辑和验证关系
- [itrace](./interfaces/itrace) 定义了`opentracing`模块提供的接口以及对各模块使用的标签定义
- [iworker](./interfaces/iworker) 定义了定时器运行队列的调用接口

## [kits](./kits) 为静态工具库

