# APM对比

## jaeger 

- 语言：go
- 背景：uber, `CNCF基金会`
- 储存：es,cassandra
- 写入性能：依赖于db/kafka
- 架构：client => (agent/sidecar=>) collector => (kafka=>) db => query-ui
- 集群：collector 无状态 最终状态依赖于es
- 客户端：go原生  支持多种采样模式 采样器插件化可二次开发 多种上报模式http/udp/grpc  `opentracing`API
- 特性：时差自适应校正 tag查询 自带`prometheus`监控和日志生态 有`all-in-one`镜像 方便多环境调试
- 已知问题: 
    1. 缺少聚合功能 
    2. 服务拓扑图较和UI较简陋


## skywalking

- 语言：java
- 背景：个人(华为工程师), `Apache基金会`
- 储存：mysql,es,TiDB,H2, InfluxDB
- 写入性能：依赖于db
- 架构：client => skywalking-oap => db => skywalking-ui
- 集群：需依赖`zookeeper`等第三方集群管理器
- 客户端：`go2sky`非官方维护  无采样率 只支持grpc上报 
- 特性：丰富的聚合分析展示 自定义api
- 已知问题：
    1. 没有时差校正 多机器采样会有偏移
    2. 没有tag查询 
    3. go端没有采样率控制(go2sky采样率实现不靠谱) 容易写爆es 
    4. 依赖es索引完整 易产生写入错误 
    5. 查询精度较差 
    6. 客户端没有实现`opentracing`API 无法适应社区生态  
    7. 多次finish `span`会导致程序崩溃