package itrace

// 所有定义
// https://github.com/apache/skywalking/blob/master/oap-server/server-bootstrap/src/main/resources/component-libraries.yml

const (
	ComponentUnknown              = 0
	ComponentTomcat               = 1
	ComponentHttpClient           = 2
	ComponentDubbo                = 3
	ComponentH2                   = 4
	ComponentMysql                = 5
	ComponentORACLE               = 6
	ComponentRedis                = 7
	ComponentMotan                = 8
	ComponentMongoDB              = 9
	ComponentResin                = 10
	ComponentFeign                = 11
	ComponentOKHttp               = 12
	ComponentSpringRestTemplate   = 13
	ComponentSpringMVC            = 14
	ComponentStruts2              = 15
	ComponentNutzMVC              = 16
	ComponentNutzHttp             = 17
	ComponentJettyClient          = 18
	ComponentJettyServer          = 19
	ComponentMemcached            = 20
	ComponentShardingJDBC         = 21
	ComponentPostgreSQL           = 22
	ComponentGRPC                 = 23
	ComponentElasticJob           = 24
	ComponentRocketMQ             = 25
	ComponentHttpasyncclient      = 26
	ComponentKafka                = 27
	ComponentServiceComb          = 28
	ComponentHystrix              = 29
	ComponentJedis                = 30
	ComponentSQLite               = 31
	ComponentH2JdbcDriver         = 32
	ComponentMysqlConnectorJava   = 33
	ComponentOjdbc                = 34
	ComponentSpymemcached         = 35
	ComponentXmemcached           = 36
	ComponentPostgresqlJdbcDriver = 37
	ComponentRocketMQProducer     = 38
	ComponentRocketMQConsumer     = 39
	ComponentKafkaProducer        = 40
	ComponentKafkaConsumer        = 41
	ComponentMongodbDriver        = 42
	ComponentSOFARPC              = 43
	ComponentActiveMQ             = 44
	ComponentActivemqProducer     = 45
	ComponentActivemqConsumer     = 46
	ComponentElasticsearch        = 47
	ComponentTransportClient      = 48
	ComponentHttp                 = 49
	ComponentRpc                  = 50
	ComponentRabbitMQ             = 51
	ComponentRabbitmqProducer     = 52
	ComponentRabbitmqConsumer     = 53
	ComponentCanal                = 54
	ComponentGson                 = 55
	ComponentRedisson             = 56
	ComponentLettuce              = 57
	ComponentZookeeper            = 58
	ComponentVertx                = 59
	ComponentShardingSphere       = 60
	ComponentSpringCloudGateway   = 61
	ComponentRESTEasy             = 62
	ComponentSolrJ                = 63
	ComponentSolr                 = 64
	ComponentSpringAsync          = 65
	ComponentJdkHttp              = 66
	ComponentSpringWebflux        = 67
	ComponentPlay                 = 68
	ComponentCassandraJavaDriver  = 69
	ComponentCassandra            = 70
	ComponentLight4J              = 71
	ComponentPulsar               = 72
	ComponentPulsarProducer       = 73
	ComponentPulsarConsumer       = 74
	ComponentEhcache              = 75
	ComponentSocketIO             = 76
	ComponentRestHighLevelClient  = 77
	ComponentSpringTx             = 78
	ComponentArmeria              = 79
	ComponentJdkThreading         = 80
	ComponentKotlinCoroutine      = 81
	ComponentAvroServer           = 82
	ComponentAvroClient           = 83
	ComponentUndertow             = 84
	ComponentFinagle              = 85
	ComponentMariadb              = 86
	ComponentMariadbJdbc          = 87
	ComponentQuasar               = 88
	ComponentInfluxDB             = 89
	ComponentInfluxdbJava         = 90
	ComponentBrpcJava             = 91
	ComponentGraphQL              = 92

	// C# .NET
	ComponentAspNetCore                          = 3001
	ComponentEntityFrameworkCore                 = 3002
	ComponentSqlClient                           = 3003
	ComponentCAP                                 = 3004
	ComponentStackExchangeRedis                  = 3005
	ComponentSqlServer                           = 3006
	ComponentNpgsql                              = 3007
	ComponentMySqlConnector                      = 3008
	ComponentEntityFrameworkCoreInMemory         = 3009
	ComponentEntityFrameworkCoreSqlServer        = 3010
	ComponentEntityFrameworkCoreSqlite           = 3011
	ComponentPomeloEntityFrameworkCoreMySql      = 3012
	ComponentNpgsqlEntityFrameworkCorePostgreSQL = 3013
	ComponentInMemoryDatabase                    = 3014
	ComponentAspNet                              = 3015
	ComponentSmartSql                            = 3016

	// NoeJS components
	// [4000, 5000) for Node.js agent
	ComponentHttpServer = 4001
	ComponentExpress    = 4002
	ComponentEgg        = 4003
	ComponentKoa        = 4004

	// Golang components
	// [5000, 6000) for Golang agent
	ComponentServiceCombMesher        = 5001
	ComponentServiceCombServiceCenter = 5002
	ComponentMOSN                     = 5003
	ComponentGoHttpServer             = 5004
	ComponentGoHttpClient             = 5005
	ComponentGin                      = 5006
	ComponentGear                     = 5007

	// Lua
	ComponentNginx = 6000

	ComponentPython   = 7000
	ComponentFlask    = 7001
	ComponentRequests = 7002

	//PHP
	ComponentPHP    = 8001
	ComponentCURL   = 8002
	ComponentPDO    = 8003
	ComponentMysqli = 8004
	ComponentYar    = 8005
	ComponentPredis = 8006
)
