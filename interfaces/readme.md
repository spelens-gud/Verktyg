# 基础组件抽象接口

此包内所有package需以i开头 代表接口(interface)

- [iconfig](./iconfig) 定义了配置加载层的接口 包括如何以什么参数获取配置 以及部分应用功能环境变量 
- [idoc](./idoc) 参照`Openapi`定义了接口文档规范 包括 `Json schema`的相关定义 后续应被官方定义替代
- [ierror](./ierror) 定义了错误类型以及错误码对http服务状态码的映射 参考`grpc`定义
- [ihttp](./ihttp) 定义了http调用的接口 以及重定义了系统默认的http调用client
- [ikafka](./ikafka) 定义了基于`sarama`的`kafka`调用接口和对象构造关系
- [ilog](./ilog) 定义了日志模块的规范 包括部分Tag以及日志级别等
- [imetrics](./imetrics) 定义了聚合监控的调用规范 以及部分命名空间 系统划分
- [iredis](./iredis) 定义了`redis`调用的规范 包括构造时需要传入的参数 以及如何提供依赖
- [isql](./isql) 定义了`sql`调用的规范 包括构造时需要传入的参数 以及如何提供依赖
- [istore](./istore) 定义了一个`KVStore`的抽象层 以及构造参数
- [itest](./itest) 定义了http接口测试的逻辑和验证关系
- [itrace](./itrace) 定义了`opentracing`模块提供的接口以及对各模块使用的标签定义
- [iworker](./iworker) 定义了定时器运行队列的调用接口