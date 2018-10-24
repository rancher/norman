// +build k3s

package k8s

import (
	"context"
	"os"

	"github.com/rancher/norman/pkg/kwrapper/etcd"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/pkg/wrapper/server"
)

func getEmbedded(ctx context.Context) (bool, context.Context, *rest.Config, error) {
	sc, ok := ctx.Value(serverConfig).(*server.ServerConfig)
	if !ok {
		ctx, sc, _, err = NewK3sConfig(ctx, "./k3s", nil)
		if err != nil {
			return false, ctx, nil, err
		}
		sc.NoScheduler = false
	}

	if len(sc.ETCDEndpoints) == 0 {
		etcdEndpoints, err := etcd.RunETCD(ctx)
		if err != nil {
			return ctx, nil, nil, err
		}
		sc.ETCDEndpoints = etcdEndpoints
	}

	err := server.Server(ctx, sc)
	if err != nil {
		return false, ctx, nil, err
	}

	os.Setenv("KUBECONFIG", sc.KubeConfig)
	restConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: sc.KubeConfig}, &clientcmd.ConfigOverrides{}).ClientConfig()

	return true, ctx, restConfig, err
}
