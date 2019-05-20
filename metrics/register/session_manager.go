package register

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rancher/norman/metrics"
)

func init() {
	prometheus.MustRegister(metrics.TotalAddWS)
	prometheus.MustRegister(metrics.TotalRemoveWS)
	prometheus.MustRegister(metrics.TotalAddConnectionsForWS)
	prometheus.MustRegister(metrics.TotalRemoveConnectionsForWS)
	prometheus.MustRegister(metrics.TotalTransmitBytesOnWS)
	prometheus.MustRegister(metrics.TotalTransmitErrorBytesOnWS)
	prometheus.MustRegister(metrics.TotalReceiveBytesOnWS)
	prometheus.MustRegister(metrics.TotalAddPeerAttempt)
	prometheus.MustRegister(metrics.TotalPeerConnected)
	prometheus.MustRegister(metrics.TotalPeerDisConnected)
}
