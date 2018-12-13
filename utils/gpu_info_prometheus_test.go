package utils

import (
	"fmt"
	"github.com/xieydd/gpu-metric/base"
	"strings"
	"testing"
)

func TestQueryMetricByPrometheus(t *testing.T) {
	clientset := base.GetClientSetForTest(t)
	if clientset == nil {
		t.Skip("kubeclient not setup")
	}
	gpuMetrics, _ := QueryMetricByPrometheus(clientset, "prometheus-svc", fmt.Sprintf(`{__name__=~"%s"}`, strings.Join(GPU_METRIC_LIST, "|")))
	for _, m := range gpuMetrics {
		t.Logf("metric %++v", m)
		t.Logf("metric name %s, value: %s", m.MetricName, m.Value)
	}
}
