package iconfig

const (
	Development = "Development" // 开发
	Testing     = "Testing"     // 测试
	PreRelease  = "PreRelease"  // 预发布
	Production  = "Production"  // 正式

	// 应用 相关配置
	envKeyOld = "env" // Deprecated
	EnvKey    = "ENV"

	// 配置加载相关配置
	EnvKeyAcmDataID    = "ACM_DATA_ID"
	EnvKeyAcmGroupID   = "ACM_GROUP_ID"
	EnvKeyAcmTenantID  = "ACM_TENANT_ID"
	EnvKeyAcmAccessID  = "ACM_ACCESS_ID"
	EnvKeyAcmAccessKey = "ACM_ACCESS_KEY"
	EnvKeyAcmEndpoint  = "ACM_ENDPOINT"

	// 容器 相关配置
	EnvKeyContainerAppName   = "CONTAINER_APP_NAME"
	EnvKeyContainerNamespace = "CONTAINER_NAMESPACE"
	EnvKeyContainerNodeName  = "CONTAINER_NODE_NAME"

	// 其他 相关配置
	EnvKeyLogLevel              = "LOG_LEVEL"                // 日志级别 默认INFO
	EnvKeyLogFileDir            = "LOG_FILE_DIR"             // 日志打印目录，默认不输出到文件
	EnvKeyLogDisableRequestBody = "LOG_DISABLE_REQUEST_BODY" // 服务器日志不输出请求体  默认开启
	EnvKeyLogEnableSql          = "LOG_ENABLE_SQL"           // Deprecated: 已废弃 默认开启日志
	EnvKeyLogDisableSqlGorm     = "LOG_DISABLE_SQL_GORM"     // 关闭gorm日志 默认开启
	EnvKeyLogDisableSqlError    = "LOG_DISABLE_SQL_ERROR"    // 关闭sql错误日志  默认开启
	EnvKeyLogDisableRedisError  = "LOG_DISABLE_REDIS_ERROR"  // 关闭redis错误日志  默认开启
	EnvKeyLogDisableAsync       = "LOG_DISABLE_ASYNC"        // 关闭日志异步输出  默认开启
	EnvKeyLogDisableFileMutex   = "LOG_DISABLE_FILE_MUTEX"   // 关闭日志文件IO全局锁 默认开启

	EnvKeyServerSentinelCpuTrigger = "SERVER_SENTINEL_CPU_TRIGGER" // 触发sentinel限流的cpu阈值 默认 0.85
	EnvKeyServerSentinelDisableBBR = "SERVER_SENTINEL_DISABLE_BBR" // 触发sentinel限流禁用BBR
	EnvKeyServerSentinelDisable    = "SERVER_SENTINEL_DISABLE"     // 中间件禁用sentinel
	EnvKeyServerTraceDisable       = "SERVER_TRACE_DISABLE"        // 中间件禁用链路追踪
	EnvKeyServerMetricsDisable     = "SERVER_METRICS_DISABLE"      // 关闭服务器监控采样
	// 监控采样
	EnvKeyServerGracefulReload   = "SERVER_ENABLE_GRACEFUL_RELOAD" // 开启优雅重启 默认关闭
	EnvKeyServerCloseWaitSeconds = "SERVER_CLOSE_WAIT"             // 服务关闭等待 单位 秒 默认20秒
	EnvKeyServerHealthCheckRoute = "SERVER_HEALTH_CHECK_ROUTE"     // 服务健康检查路径 该路径不会被记录链路追踪日志等 默认 "HEAD:/"
	EnvKeyServerEnableOKLogOmit  = "SERVER_ENABLE_OK_LOG_OMIT"     // 开启此选项后 所有2XX成功请求都不会输出日志

	// Deprecated: 异步日志配置和全局日志开关同步
	// EnvKeyServerEnableLogAsync      = "SERVER_ENABLE_LOG_ASYNC"

	// Deprecated: 日志输出不使用全局锁 使用 LOG_DISABLE_FILE_MUTEX
	// EnvKeyServerDisableLogMutex     = "SERVER_DISABLE_LOG_MUTEX"

	EnvKeyServerCloseWaitSecondsOld = "CLOSE_WAIT"           // Deprecated:服务关闭等待旧版
	EnvKeyGovernServerEnable        = "GOVERN_SERVER_ENABLE" // 开启扩展服务 默认关闭
	EnvKeyGovernServerPort          = "GOVERN_SERVER_PORT"

	EnvKeyRuntimePprofEnable            = "RUNTIME_PPROF_ENABLE"       // 开启启用pprof 默认关闭
	EnvKeyRuntimePprofPrefix            = "RUNTIME_PPROF_PREFIX"       // pprof访问路径前缀 默认 /debug/pprof
	EnvKeyRuntimePprofSecretKey         = "RUNTIME_PPROF_SECRET_KEY"   // pprof访问密钥 默认 空
	EnvKeyRuntimeMetricsEnable          = "RUNTIME_METRICS_ENABLE"     // Deprecated: 开启启用metrics直接访问路径 废弃 metrics访问路径不为空即暴露
	EnvKeyRuntimeMetricsExportSecretKey = "RUNTIME_METRICS_SECRET_KEY" // metrics访问密钥 默认 空
	EnvKeyRuntimeMetricsExportPath      = "RUNTIME_METRICS_PATH"       // metrics访问路径 默认 /metrics

	// Deprecated: 此环境变量弃用，默认关闭请求client路径metrics路径采集
	// EnvKeyRuntimeMetricsDisableClientPath = "RUNTIME_METRICS_DISABLE_CLIENT_PATH" // 是否禁用请求client路径metrics 默认开启

	EnvKeyRuntimeMetricsEnableClientPath = "RUNTIME_METRICS_ENABLE_CLIENT_PATH" // 是否开启请求client路径metrics采集 默认关闭

	EnvKeyTracerEnable = "TRACER_ENABLE" // 使用的链路追踪器 默认 jaeger 填 False/Disable 则关闭链路追踪
)
