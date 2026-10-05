package k8s

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

type ContainerUsage struct {
	Namespace   string
	Pod         string
	Container   string
	MemoryBytes int64
}

func NewMetricsClient(opts Options) (*metricsclient.Clientset, error) {
	cfg, err := BuildConfig(opts)
	if err != nil {
		return nil, err
	}
	cfg.WarningHandler = rest.NoWarnings{}
	return metricsclient.NewForConfig(cfg)
}

func CollectMetrics(ctx context.Context, mc metricsclient.Interface, namespace string, snap *Snapshot) {
	if mc == nil {
		snap.Degraded = append(snap.Degraded, "pod usage unavailable (no metrics client)")
		return
	}
	res, err := mc.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{Limit: 500})
	if err != nil {
		snap.Degraded = append(snap.Degraded, "pod usage unavailable ("+err.Error()+")")
		return
	}
	for _, pm := range res.Items {
		for _, c := range pm.Containers {
			q := c.Usage.Memory()
			snap.Usage = append(snap.Usage, ContainerUsage{
				Namespace:   pm.Namespace,
				Pod:         pm.Name,
				Container:   c.Name,
				MemoryBytes: q.Value(),
			})
		}
	}
}
