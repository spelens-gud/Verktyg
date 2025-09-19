package imetrics

import "github.com/prometheus/client_golang/prometheus"

type desc struct {
	desc         *prometheus.Desc
	metricCaller func(*prometheus.Desc, ...interface{}) prometheus.Metric
}

type DescGroup struct {
	group           map[string]desc
	constLabels     map[string]string
	namespace       string
	subSystem       string
	collectCallback func() []interface{}
}

func NewDescGroup(namespace, subSystem string, constLabels map[string]string, collectCallback func() []interface{}) *DescGroup {
	return &DescGroup{
		group:           make(map[string]desc),
		constLabels:     constLabels,
		namespace:       namespace,
		subSystem:       subSystem,
		collectCallback: collectCallback,
	}
}

func (g *DescGroup) Describe(ch chan<- *prometheus.Desc) {
	for _, d := range g.group {
		ch <- d.desc
	}
}

func (g *DescGroup) Collect(ch chan<- prometheus.Metric) {
	for _, d := range g.group {
		ch <- d.metricCaller(d.desc, g.collectCallback()...)
	}
}

func (g *DescGroup) NewConstMetrics(name string, help string, typ prometheus.ValueType, metrics func(...interface{}) float64, labels ...string) {
	g.New(name, help, func(p *prometheus.Desc, items ...interface{}) prometheus.Metric {
		return prometheus.MustNewConstMetric(p, typ, metrics(items...))
	}, labels...)
}

func (g *DescGroup) New(name string, help string, metrics func(*prometheus.Desc, ...interface{}) prometheus.Metric, labels ...string) {
	g.group[name] = desc{
		desc:         prometheus.NewDesc(prometheus.BuildFQName(g.namespace, g.subSystem, name), help, labels, g.constLabels),
		metricCaller: metrics,
	}
}
