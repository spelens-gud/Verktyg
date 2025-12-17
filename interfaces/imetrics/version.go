package imetrics

import (
	prom "github.com/prometheus/client_golang/prometheus"

	"github.com/spelens-gud/Verktyg/version"
)

var versionDesc = prom.NewDesc(
	"go_mod_version_sy_core",
	"Version of sy-core go mod dependencies version",
	nil, prom.Labels{"version": version.GetVersion()})

type SYCoreVersion struct{}

func (s SYCoreVersion) Describe(ch chan<- *prom.Desc) {
	ch <- versionDesc
}

func (s SYCoreVersion) Collect(ch chan<- prom.Metric) {
	ch <- prom.MustNewConstMetric(versionDesc, prom.GaugeValue, 1)
}

var _ prom.Collector = &SYCoreVersion{}
