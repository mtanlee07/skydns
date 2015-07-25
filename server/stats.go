// Copyright (c) 2014 The SkyDNS Authors. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package server

import (
	"net/http"
	"os"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	prometheusPort      = os.Getenv("PROMETHEUS_PORT")
	prometheusPath      = os.Getenv("PROMETHEUS_PATH")
	prometheusNamespace = os.Getenv("PROMETHEUS_NAMESPACE")
	prometheusSubsystem = os.Getenv("PROMETHEUS_SUBSYSTEM")
)

var (
	promDnssecOkCount        prometheus.Counter
	promExternalRequestCount *prometheus.CounterVec
	promRequestCount         *prometheus.CounterVec
	promErrorCount           *prometheus.CounterVec
	promBackendFailureCount  *prometheus.CounterVec
	promCacheSize            *prometheus.GaugeVec
	promCacheMiss            *prometheus.CounterVec
	promRequestDuration      *prometheus.HistogramVec
	promResponseSize         *prometheus.HistogramVec
)

func Metrics() {
	if prometheusPath == "" {
		prometheusPath = "/metrics"
	}
	if prometheusSubsystem == "" {
		prometheusSubsystem = "skydns"
	}

	promExternalRequestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: prometheusNamespace,
		Subsystem: prometheusSubsystem,
		Name:      "dns_request_external_count",
		Help:      "Counter of external DNS requests.",
	}, []string{"type"}) // recursive, stub, lookup
	prometheus.MustRegister(promExternalRequestCount)

	promRequestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: prometheusNamespace,
		Subsystem: prometheusSubsystem,
		Name:      "dns_request_count",
		Help:      "Counter of DNS requests made.",
	}, []string{"type"}) // udp, tcp
	prometheus.MustRegister(promRequestCount)

	promDnssecOkCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: prometheusNamespace,
		Subsystem: prometheusSubsystem,
		Name:      "dns_dnssec_ok_count",
		Help:      "Counter of DNSSEC requests.",
	})
	prometheus.MustRegister(promDnssecOkCount) // Maybe more bits here?

	promBackendFailureCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: prometheusNamespace,
		Subsystem: prometheusSubsystem,
		Name:      "backend_failure_count",
		Help:      "Counter of JSON parsing failures.",
	}, []string{"type"}) // other, etcd (etcd not used at the moment)
	prometheus.MustRegister(promBackendFailureCount)

	promErrorCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: prometheusNamespace,
		Subsystem: prometheusSubsystem,
		Name:      "dns_error_count",
		Help:      "Counter of DNS requests resulting in an error.",
	}, []string{"error"}) // nxdomain, nodata, truncated, refused, overflow
	prometheus.MustRegister(promErrorCount)

	// Caches
	promCacheSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: prometheusNamespace,
		Subsystem: prometheusSubsystem,
		Name:      "cache_total_size",
		Help:      "The total size of all elements in the cache.",
	}, []string{"type"}) // response, signature
	prometheus.MustRegister(promCacheSize)

	promCacheMiss = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: prometheusNamespace,
		Subsystem: prometheusSubsystem,
		Name:      "dns_cache_miss_count",
		Help:      "Counter of DNS requests that result in a cache miss.",
	}, []string{"type"}) // response, signature
	prometheus.MustRegister(promCacheMiss)

	promRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: prometheusNamespace,
		Subsystem: prometheusSubsystem,
		Name:      "dns_request_duration",
		Help:      "Histogram of the time (in seconds) each request took to resolve.",
		Buckets:   append([]float64{0.001, 0.003}, prometheus.DefBuckets...),
	}, []string{"type"}) // udp, tcp
	prometheus.MustRegister(promRequestDuration)

	promResponseSize = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: prometheusNamespace,
		Subsystem: prometheusSubsystem,
		Name:      "dns_response_size",
		Help:      "Size of the returns response in bytes.",
		// Powers of 2 up to the maximum size.
		Buckets: []float64{0, 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768, 65536},
	}, []string{"type"}) // udp, tcp
	prometheus.MustRegister(promResponseSize)

	if prometheusPort == "" {
		return
	}

	_, err := strconv.Atoi(prometheusPort)
	if err != nil {
		Fatalf("bad port for prometheus: %s", prometheusPort)
	}

	http.Handle(prometheusPath, prometheus.Handler())
	go func() {
		Fatalf("%s", http.ListenAndServe(":"+prometheusPort, nil))
	}()
	Logf("metrics enabled on :%s%s", prometheusPort, prometheusPath)
}

// Counter is the metric interface used by this package
type Counter interface {
	Inc(i int64)
}

type nopCounter struct{}

func (nopCounter) Inc(_ int64) {}

// These are the old stat variables defined by this package. This
// used by graphite.
var (
	// Pondering deletion in favor of the better and more
	// maintained (by me) prometheus reporting.

	StatsForwardCount     Counter = nopCounter{}
	StatsStubForwardCount Counter = nopCounter{}
	StatsLookupCount      Counter = nopCounter{}
	StatsRequestCount     Counter = nopCounter{}
	StatsDnssecOkCount    Counter = nopCounter{}
	StatsNameErrorCount   Counter = nopCounter{}
	StatsNoDataCount      Counter = nopCounter{}

	StatsDnssecCacheMiss Counter = nopCounter{}
)
