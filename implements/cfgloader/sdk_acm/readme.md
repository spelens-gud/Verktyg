# ACM/SCM客户端

- ACM提供应用配置管理
- SCM提供业务配置管理

```go
    // 新建client 指定调用地址
    var cli = sdk_acm.NewClient("http://scm-dev.39on.com")

	// 获取配置内容
	content, err := cli.GetConfig(context.Background(), sdk_acm.ConfigData{
		TenantID: "", // 集群ID
		DataID:   "", // 数据ID
		GroupID:  "", // 组ID
	})
	// 更新配置内容
	content, err := cli.UpdateConfig(context.Background(), sdk_acm.ConfigData{
		TenantID: "", // 集群ID
		DataID:   "", // 数据ID
		GroupID:  "", // 组ID
		Content: "", // 内容
	})
	// 更新配置组数据ID集合 清除不存在的数据ID
	err = cli.UpdateConfigSet(context.Background(), sdk_acm.ConfigSet{
		GroupID:  "",         // 组ID
		DataIDs:  []string{}, // 数据组内的数据ID集合
		TenantID: "",         // 集群ID
	})
```