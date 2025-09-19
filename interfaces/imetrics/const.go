package imetrics

const (
	NamespaceSqlStats   = "go_sql_stats"
	NamespaceSql        = "sql"
	NamespaceRedisStats = "go_redis_stats"
	NamespaceRedis      = "redis"
	NamespaceMq         = "mq"
	NamespaceHttp       = "http"
	NamespaceRpc        = "rpc"

	NameHandlingSeconds = "handling_seconds"
	NameDurationSeconds = "duration_s"

	SubsystemConnections = "connections"
	SubsystemUsage       = "usage"
	SubsystemClient      = "client"
	SubsystemServer      = "server"
)

var DefBucket = []float64{0.0005, 0.001, 0.0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
