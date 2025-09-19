package imetrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	counter   *prometheus.CounterVec
	histogram *prometheus.HistogramVec
}

type MetricsGroup struct {
	group       map[string]metrics
	constLabels map[string]string
	labels      map[string][]string
	namespace   string
	subSystem   string
}

func NewMetricsGroup(namespace, subSystem string, constLabels map[string]string) *MetricsGroup {
	return &MetricsGroup{
		group:       make(map[string]metrics),
		labels:      make(map[string][]string),
		constLabels: constLabels,
		namespace:   namespace,
		subSystem:   subSystem,
	}
}

func (g *MetricsGroup) Describe(ch chan<- *prometheus.Desc) {
	for _, d := range g.group {
		if d.counter != nil {
			d.counter.Describe(ch)
		} else if d.histogram != nil {
			d.histogram.Describe(ch)
		}
	}
}

func (g *MetricsGroup) Collect(ch chan<- prometheus.Metric) {
	for _, d := range g.group {
		if d.counter != nil {
			d.counter.Collect(ch)
		} else if d.histogram != nil {
			d.histogram.Collect(ch)
		}
	}
}

func (g *MetricsGroup) NewCounter(name string, help string, labels ...string) {
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   g.namespace,
		Subsystem:   g.subSystem,
		Name:        name,
		Help:        help,
		ConstLabels: g.constLabels,
	}, labels)

	g.group[name] = metrics{
		counter: counter,
	}
	g.labels[name] = labels
}

func (g *MetricsGroup) NewHistogram(name string, help string, labels ...string) {
	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   g.namespace,
		Subsystem:   g.subSystem,
		Name:        name,
		Help:        help,
		ConstLabels: g.constLabels,
		Buckets:     DefBucket,
	}, labels)

	g.group[name] = metrics{
		histogram: histogram,
	}
	g.labels[name] = labels
}

func (g *MetricsGroup) AddX(name string, count float64, labelMaps ...map[string]string) {
	labels := make([]string, len(g.labels[name]))
	for i := range labels {
		for _, labelMap := range labelMaps {
			if val, ok := labelMap[g.labels[name][i]]; ok {
				labels[i] = val
			}
		}
	}
	g.Add(name, count, labels...)
}

func (g *MetricsGroup) Add(name string, count float64, labels ...string) {
	counter, ok := g.group[name]
	if !ok {
		return
	}

	labelCount := len(g.labels[name])
	labelCount2Set := len(labels)

	switch {
	case labelCount2Set > labelCount:
		labels = labels[:labelCount]
	case labelCount2Set < labelCount:
		for i := 0; i < labelCount-labelCount2Set; i++ {
			labels = append(labels, "")
		}
	}

	if counter.counter != nil {
		counter.counter.WithLabelValues(labels...).Add(count)
	}
	if counter.histogram != nil {
		counter.histogram.WithLabelValues(labels...).Observe(count)
	}
}
