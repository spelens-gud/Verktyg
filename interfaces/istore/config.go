package istore

import (
	"errors"
	"strconv"
	"strings"
)

var RegisteredStore = make(map[string]func(servers ...string) (Store, error))

const DefaultStoreType = "memory"

type ConfigStore struct {
	Servers       []string      `json:"servers"`         // 服务地址 若提供固定地址 直接使用地址
	StartPort     int           `json:"start_port"`      // 起始端口 使用地址格式和端口自动生成地址
	PortNum       int           `json:"port_num"`        // 端口总数
	Addr          string        `json:"addr"`            // 地址格式
	PortIdent     string        `json:"port_ident"`      // 地址格式的端口格式字符 默认为 {{port}}
	EnableAutoBak bool          `json:"enable_auto_bak"` // 启动自动切换备用地址
	BakConfig     []ConfigStore `json:"bak_config"`      // 备用地址
	Type          string        `json:"type"`            // store类型 默认为memory
}

func (cfg ConfigStore) FixAddr() []string {
	if len(cfg.PortIdent) == 0 {
		cfg.PortIdent = "{{port}}"
	}
	if len(cfg.Servers) == 0 && cfg.StartPort > 0 && cfg.PortNum > 0 && len(cfg.Addr) > 0 {
		for port := cfg.StartPort; port < cfg.PortNum+cfg.StartPort; port++ {
			addr := strings.ReplaceAll(cfg.Addr, cfg.PortIdent, strconv.Itoa(port))
			cfg.Servers = append(cfg.Servers, addr)
		}
	}
	return cfg.Servers
}

func (cfg ConfigStore) NewStore() (st Store, err error) {
	if len(cfg.Type) == 0 {
		cfg.Type = DefaultStoreType
	}
	fac, ok := RegisteredStore[cfg.Type]
	if ok {
		return cfg.NewStoreFromFactory(fac)
	}
	err = errors.New("invalid store type")
	return
}

func (cfg ConfigStore) NewStoreFromFactory(factoryFunc func(addr ...string) (Store, error)) (Store, error) {
	cfg.Servers = cfg.FixAddr()
	if len(cfg.Servers) == 0 {
		err := errors.New("empty servers")
		return nil, err
	}

	st, err := factoryFunc(cfg.Servers...)
	if err != nil {
		return nil, err
	}

	if cfg.EnableAutoBak && len(cfg.BakConfig) > 0 {
		baks := make([]Store, 0, len(cfg.BakConfig))
		for _, bf := range cfg.BakConfig {
			bf.Servers = bf.FixAddr()
			if len(bf.Servers) == 0 {
				err := errors.New("empty bak servers")
				return nil, err
			}
			bak, err := factoryFunc(bf.Servers...)
			if err != nil {
				return nil, err
			}
			baks = append(baks, bak)
		}
		return NewBakStore(st, baks), nil
	}
	return st, nil
}
