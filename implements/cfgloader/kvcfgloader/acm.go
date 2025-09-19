package kvcfgloader

import (
	"context"

	"git.bestfulfill.tech/devops/go-core/implements/cfgloader/sdk_acm"
	"git.bestfulfill.tech/devops/go-core/interfaces/iconfig"
)

// 应用KV配置
type AcmKVConfig struct {
	tenantID string
	cli      sdk_acm.Client

	opt acmKVOption
}

func NewAcmKVConfig(tenantID string, cli sdk_acm.Client, opts ...AcmKVOption) *AcmKVConfig {
	o := AcmKVOptions(opts).Init()
	return &AcmKVConfig{
		tenantID: tenantID,
		cli:      cli,
		opt:      *o,
	}
}

func (a *AcmKVConfig) GetWithContext(ctx context.Context, key string) (value string, err error) {
	groupID, dataID := a.opt.keySplit(key)
	ret, err := a.cli.GetOriginConfig(ctx, sdk_acm.ConfigData{
		DataID:   dataID,
		GroupID:  groupID,
		TenantID: a.tenantID,
	})
	if err != nil {
		return
	}
	value = string(ret)
	return
}

func (a *AcmKVConfig) SetWithContext(ctx context.Context, key string, value string) (err error) {
	dataID, groupID := a.opt.keySplit(key)
	err = a.cli.UpdateConfig(ctx, sdk_acm.ConfigData{
		DataID:   dataID,
		GroupID:  groupID,
		Content:  value,
		TenantID: a.tenantID,
	})
	return
}

var _ iconfig.KVConfigSourceLoader = &AcmKVConfig{}
