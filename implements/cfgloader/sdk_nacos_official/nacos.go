package sdk_nacos_official

import (
	"errors"
	"net/url"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"

	"github.com/spelens-gud/Verktyg.git/interfaces/ihttp"
	"github.com/spelens-gud/Verktyg.git/kits/kjson"
)

type Client interface {
	config_client.IConfigClient

	GetRefConfig(param vo.ConfigParam) (content []byte, err error)
}

func NewClient(endpoint, tenantID, cacheDir string) (cli Client, err error) {
	var sc []constant.ServerConfig
	sp := strings.Split(endpoint, " ")
	for _, s := range sp {
		u, e := url.Parse(s)
		if e != nil {
			continue
		}
		host, port, e := ihttp.SplitHostPort(u.Host)
		if e != nil {
			continue
		}
		sc = append(sc,
			constant.ServerConfig{
				IpAddr: host,
				Port:   uint64(port),
			})
	}
	if len(sp) == 0 {
		err = errors.New("invalid endpoint")
		return
	}
	cc := constant.ClientConfig{
		NamespaceId:         tenantID,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		CacheDir:            cacheDir,
	}

	nacosCli, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": sc,
		"clientConfig":  cc,
	})
	if err != nil {
		return
	}
	cli = client{IConfigClient: nacosCli}
	return
}

type client struct {
	config_client.IConfigClient
}

func (c client) GetRefConfig(param vo.ConfigParam) (content []byte, err error) {
	ret, err := c.GetConfig(param)
	if err != nil {
		return
	}
	content, err = kjson.LoadRefJson([]byte(ret), func(string string) (ref []byte, err error) {
		newData := param
		if strings.ContainsRune(string, '.') {
			sp := strings.Split(string, ".")
			newData.Group = sp[0]
			newData.DataId = sp[1]
		} else {
			newData.DataId = string
		}
		data, err := c.GetConfig(newData)
		ref = []byte(data)
		return
	})
	return
}
