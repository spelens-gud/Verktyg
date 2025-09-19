package promgateway

import (
	"context"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"

	"git.bestfulfill.tech/devops/go-core/implements/httpreq"
	"git.bestfulfill.tech/devops/go-core/interfaces/iconfig"
	"git.bestfulfill.tech/devops/go-core/interfaces/ihttp"
	"git.bestfulfill.tech/devops/go-core/interfaces/imetrics"
	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
)

type (
	Transport struct {
		pusher *push.Pusher

		client *http.Client

		startOnce  sync.Once
		cancelFunc func()
		closed     chan struct{}

		config *GatewayConfig
	}

	Options func(*GatewayConfig)

	GatewayConfig struct {
		Job                string `json:"job"`
		GatewayUrl         string `json:"gateway_url"`
		IntervalSeconds    int    `json:"interval_seconds"`
		EnableLog          bool   `json:"enable_log"`
		DisableDeferDelete bool   `json:"disable_defer_delete"`
		BasicAuth

		GroupingName string                 `json:"-"`
		GroupingKV   []func() (k, v string) `json:"-"`
		ReqOpt       []ihttp.ReqOpt         `json:"-"`
	}

	BasicAuth struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
)

const (
	defaultPushIntervalSeconds = 5
	defaultGroupingName        = "nodename"

	tagErrLog = "PROM_PUSH_GATEWAY_ERR"
)

func (cfg GatewayConfig) NewTransport(opts ...Options) imetrics.GatewayDaemon {
	for _, opt := range opts {
		opt(&cfg)
	}
	return NewGateWayTransport(cfg.GatewayUrl, cfg.Job, opts...)
}

func NewGateWayTransport(gatewayUrl, job string, opts ...Options) imetrics.GatewayDaemon {
	if len(gatewayUrl) == 0 {
		logger.FromBackground().WithTag(tagErrLog).Errorf("gateway url empty,init noop daemon")
		return NoopDaemon{}
	}

	if len(job) == 0 {
		job = iconfig.GetApplicationName()
	}

	gw := &Transport{
		config: &GatewayConfig{
			IntervalSeconds: defaultPushIntervalSeconds,
			GroupingName:    defaultGroupingName,
		},
	}

	for _, opt := range opts {
		opt(gw.config)
	}

	httpDoOpt := &httpreq.HttpDoOption{
		DisableTrace:  true,
		DisableLog:    !gw.config.EnableLog,
		MaxRetryTimes: 0,
	}

	gw.config.ReqOpt = append(gw.config.ReqOpt, httpreq.WithHttpDoOption(httpDoOpt))

	gw.client = ihttp.NewDefaultHttpClient(func(cli *http.Client) {
		cli.Timeout = time.Duration(gw.config.IntervalSeconds) * time.Second
	})

	gw.pusher = push.New(gatewayUrl, job).Client(gw)
	return gw
}

func (gw *Transport) Do(req *http.Request) (res *http.Response, err error) {
	res, err = httpreq.DoRequest(gw.client, req, gw.config.ReqOpt...)
	return
}

func (gw *Transport) Stop() {
	if gw.cancelFunc != nil {
		gw.cancelFunc()
		<-gw.closed
	}
}

func (gw *Transport) startDaemon() {
	var (
		ctx, cf  = context.WithCancel(context.Background())
		hostname = iconfig.HostName()
		interval = time.Duration(gw.config.IntervalSeconds) * time.Second
		ticker   = time.NewTicker(interval)
		lg       = logger.FromBackground()
	)

	lg.WithTag("PROM_PUSH_GATEWAY").Infof("metrics push daemon start")
	gw.cancelFunc = cf
	gw.closed = make(chan struct{})

	defer ticker.Stop()
	defer close(gw.closed)

	// hostname聚合
	// 容器拼上宿主机名称
	if nodeName := os.Getenv(iconfig.EnvKeyContainerNodeName); len(nodeName) > 0 {
		hostname = nodeName + ":" + hostname
	}
	gw.pusher.Grouping(gw.config.GroupingName, hostname)
	gw.pusher.Gatherer(prometheus.DefaultGatherer)

	// 其他实时聚合
	for _, f := range gw.config.GroupingKV {
		gw.pusher.Grouping(f())
	}

	doPush := func() {
		if err := gw.pusher.Push(); err != nil {
			lg.WithTag(tagErrLog).Errorf("push metrics gateway error: %v", err)
		}
	}

	for {
		select {
		case <-ticker.C:
			doPush()
		case <-ctx.Done():
			doPush()

			if gw.config.DisableDeferDelete {
				return
			}

			time.Sleep(interval)
			if err := gw.pusher.Delete(); err != nil {
				lg.WithTag(tagErrLog).Errorf("defer delete metrics gateway error: %v", err)
			}
			return
		}
	}
}

func (gw *Transport) StartDaemon() { gw.startOnce.Do(gw.startDaemon) }
