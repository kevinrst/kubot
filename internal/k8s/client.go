package k8s

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Options struct {
	Kubeconfig string
	Context    string
}

func ResolveContextName(opts Options) string {
	if opts.Context != "" {
		return opts.Context
	}
	loading := clientcmd.NewDefaultClientConfigLoadingRules()
	cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loading, nil)
	raw, err := cfg.RawConfig()
	if err != nil {
		return ""
	}
	return raw.CurrentContext
}

// Builds a read-only clientset from flags, env, default path, or in-cluster.
func NewClientset(opts Options) (*kubernetes.Clientset, error) {
	cfg, err := BuildConfig(opts)
	if err != nil {
		return nil, err
	}
	return NewForConfig(cfg)
}

// Resolves the kubeconfig chain into a rest.Config for alternate clients.
func BuildConfig(opts Options) (*rest.Config, error) {
	// Explicit --kubeconfig or --context.
	if opts.Kubeconfig != "" || opts.Context != "" {
		path := opts.Kubeconfig
		if path == "" {
			if env := os.Getenv("KUBECONFIG"); env != "" {
				path = env
			} else {
				path = clientcmd.NewDefaultPathOptions().GetDefaultFilename()
			}
		}
		cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: path},
			&clientcmd.ConfigOverrides{CurrentContext: opts.Context},
		).ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("load kubeconfig: %w", err)
		}
		return cfg, nil
	}
	// Default chain: kubeconfig file, then in-cluster.
	loading := clientcmd.NewDefaultClientConfigLoadingRules()
	if cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loading, nil).ClientConfig(); err == nil {
		return cfg, nil
	}
	if inCluster, err := rest.InClusterConfig(); err == nil {
		return inCluster, nil
	}
	return nil, fmt.Errorf("no cluster connection: set KUBECONFIG, pass --kubeconfig, or run inside a cluster")
}

// Drops server deprecation warnings so they never pollute kubot's output.
func NewForConfig(cfg *rest.Config) (*kubernetes.Clientset, error) {
	cfg.WarningHandler = rest.NoWarnings{}
	return kubernetes.NewForConfig(cfg)
}
