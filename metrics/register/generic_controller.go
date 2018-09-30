package register

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rancher/norman/metrics"
)

func init() {
	prometheus.MustRegister(metrics.TotalHandlerExecution)
	prometheus.MustRegister(metrics.TotalHandlerFailure)
}
