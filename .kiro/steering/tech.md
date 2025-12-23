# 技术栈

## 语言与版本

- **Go**: 1.23+
- **模块路径**: `github.com/spelens-gud/Verktyg`

## 核心依赖

### Web 框架
- `github.com/gin-gonic/gin` - HTTP 服务器框架

### 数据库
- `gorm.io/gorm` (v1.23.6) - ORM 框架
- `github.com/jinzhu/gorm` (v1.9.14) - 旧版 GORM（向后兼容）
- `github.com/go-sql-driver/mysql` - MySQL 驱动

### 缓存与存储
- `github.com/go-redis/redis/v8` - Redis 客户端
- `github.com/bradfitz/gomemcache` - Memcached 客户端
- `github.com/patrickmn/go-cache` - 内存缓存

### 消息队列
- `github.com/IBM/sarama` - Kafka 客户端
- `github.com/bsm/sarama-cluster` - Kafka 消费者组

### 配置中心
- `github.com/nacos-group/nacos-sdk-go` - Nacos SDK
- `github.com/alibaba/sentinel-golang` - 限流熔断

### 可观测性
- `github.com/prometheus/client_golang` - Prometheus 监控
- `github.com/opentracing/opentracing-go` - OpenTracing 追踪
- `github.com/uber/jaeger-client-go` - Jaeger 客户端
- `github.com/SkyAPM/go2sky` - SkyWalking 追踪
- `go.uber.org/zap` - 结构化日志

### 工具库
- `github.com/robfig/cron/v3` - 定时任务
- `github.com/spf13/cast` - 类型转换
- `github.com/pkg/errors` - 错误处理
- `golang.org/x/sync` - 并发控制

## 构建工具

### Makefile 命令

```bash
# 代码格式化和 lint 检查
make lint

# 生成 changelog（使用 git-chglog）
make cl
```

### 常用命令

```bash
# 安装依赖
go mod download

# 运行测试
go test ./...

# 构建
go build ./...

# 代码格式化
goimports -format-only -w -l -local github.com/spelens-gud/Verktyg ./

# Lint 检查
golangci-lint run ./...
```

## 开发工具

- **goimports**: 自动导入管理和格式化
- **golangci-lint**: 代码质量检查
- **git-chglog**: 自动生成 CHANGELOG
