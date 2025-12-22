# 项目结构

## 架构设计原则

本项目采用**接口-实现-工具**三层架构，遵循依赖倒置原则：

1. **interfaces/** - 定义抽象接口
2. **implements/** - 提供具体实现
3. **kits/** - 提供静态工具库

## 目录结构

### interfaces/ - 接口抽象层

定义所有组件的抽象接口，包名必须以 `i` 开头（表示 interface）。

- `iconfig/` - 配置加载接口
- `ilog/` - 日志接口
- `isql/` - 数据库接口
- `iredis/` - Redis 接口
- `ikafka/` - Kafka 接口
- `ihttp/` - HTTP 客户端接口
- `itrace/` - 链路追踪接口
- `imetrics/` - 监控指标接口
- `istore/` - KV 存储接口
- `iworker/` - 定时任务接口
- `ierror/` - 错误处理接口
- `itest/` - 测试工具接口

### implements/ - 组件实现层

实现 interfaces 中定义的接口，每个实现都以接口形式对外暴露。

- `cfgloader/` - 配置加载器（ACM、Nacos、文件）
- `zaplog/` - 基于 Zap 的日志实现
- `gormctx/` - 支持 Context 的 GORM 封装
- `gormx/` - GORM 扩展工具
- `redisx/` - Redis 客户端封装
- `skafka/` - Kafka 客户端封装
- `httpreq/` - HTTP 客户端（支持重试、熔断）
- `otrace/` - OpenTracing 实现
- `skytrace/` - SkyWalking 实现
- `promdb/` - 数据库监控
- `promhttp/` - HTTP 监控
- `promrpc/` - RPC 监控
- `promgateway/` - Prometheus PushGateway
- `store/` - KV 存储实现（Memcached、内存）
- `worker/` - 定时任务管理器
- `yapi/` - YApi 接口封装

### kits/ - 工具库层

提供静态工具函数和辅助功能，包名必须以 `k` 开头（表示 kits）。

- `kcontext/` - Context 工具（元数据、请求信息）
- `klog/` - 日志工具（文件轮转、缓冲写入）
- `kserver/` - HTTP 服务器（Gin、优雅重启、中间件）
- `kdaemon/` - 后台任务管理
- `kdb/` - 数据库初始化和监控
- `kdoc/` - 文档生成工具
- `kerror/` - 错误处理工具
- `kgo/` - Go 语言工具（字符串、分组、缓冲池）
- `kjson/` - JSON 工具
- `knet/` - 网络工具（TLS、Dialer）
- `kruntime/` - 运行时工具（信号处理）
- `kstruct/` - 结构体工具
- `ktrace/` - 追踪工具
- `kurl/` - URL 工具
- `kwire/` - Wire 依赖注入检查

### 其他目录

- `internal/` - 内部包，不对外暴露
- `version/` - 版本信息
- `docs/` - 示例代码和文档
- `scripts/` - 构建脚本

## 命名规范

### 包命名
- 接口包：以 `i` 开头（如 `ilog`, `isql`）
- 实现包：描述性名称（如 `zaplog`, `gormctx`）
- 工具包：以 `k` 开头（如 `klog`, `kserver`）

### 文件命名
- 使用小写和下划线：`http_client.go`
- 测试文件：`*_test.go`
- 示例文件：`example_*.go` 或 `*_example_test.go`

### 接口命名
- 接口名使用名词或形容词：`Logger`, `Tracer`, `Store`
- Provider 接口用于初始化：`LoggerProvider`
- Factory 接口用于构造：`RedisFactory`

## 依赖关系

```
应用代码
    ↓
interfaces (定义契约)
    ↓
implements (具体实现)
    ↓
kits (底层工具)
```

- **implements** 依赖 **interfaces**
- **kits** 不依赖 **interfaces** 和 **implements**
- 应用代码依赖 **interfaces**，运行时注入 **implements**

## 示例代码位置

- `docs/example_*/` - 各组件的使用示例
- `*_test.go` - 单元测试
- `example_*_test.go` - 可执行示例
