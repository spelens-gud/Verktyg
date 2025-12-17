package skytrace

import (
	"time"

	"github.com/SkyAPM/go2sky/reporter"

	"github.com/spelens-gud/Verktyg/interfaces/itrace"
)

type GRPCReportTracerConfig struct {
	Service          string  `json:"service"`
	Addr             string  `json:"addr"`
	Auth             string  `json:"auth"`
	CheckSeconds     int     `json:"check_seconds"`
	SampleRate       float64 `json:"sample_rate"`
	MaxSendQueueSize int     `json:"max_send_queue_size"`
}

const defaultMaxSendQueueSize = 5000

func (cfg GRPCReportTracerConfig) Names() []string {
	return []string{"skywalking", "sky"}
}

func (cfg GRPCReportTracerConfig) NewTracer() (tracer itrace.Tracer, err error) {
	opts := []reporter.GRPCReporterOption{
		reporter.WithMaxSendQueueSize(defaultMaxSendQueueSize),
	}

	if len(cfg.Auth) > 0 {
		opts = append(opts, reporter.WithAuthentication(cfg.Auth))
	}

	if cfg.MaxSendQueueSize > 0 {
		opts = append(opts, reporter.WithMaxSendQueueSize(cfg.MaxSendQueueSize))
	}

	if cfg.CheckSeconds > 0 {
		opts = append(opts, reporter.WithCheckInterval(time.Duration(cfg.CheckSeconds)*time.Second))
	}

	tracerOpts := []Option{
		func(opt *TraceOption) {
			opt.ReportOptions = append(opt.ReportOptions, opts...)
		},
	}

	if cfg.SampleRate > 0 {
		tracerOpts = append(tracerOpts, func(opt *TraceOption) {
			opt.SampleRate = cfg.SampleRate
		})
	}
	return NewGrpcTracer(cfg.Service, cfg.Addr, tracerOpts...)
}
