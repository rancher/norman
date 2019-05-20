package metrics

import (
	"os"
)

func init() {
	if os.Getenv(MetricsGenericControllerEnv) == "true" {
		genericControllerMetrics = true
	}
	if os.Getenv(MetricsSessionServerEnv) == "true" {
		sessionServerMetrics = true
	}
}
