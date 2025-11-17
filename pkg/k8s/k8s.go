package k8s

import (
	"context"
	"fake-deviceplugin/pkg/log"
	"fake-deviceplugin/pkg/utils"
	"os"
	"path/filepath"
	"sync"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeconfig string = ""
var kubeClient *clientset.Clientset
var once sync.Once

func GetKubeConfig() *rest.Config {
	// Determine kubeconfig path: explicit var -> KUBECONFIG env -> $HOME/.kube/config
	if kubeconfig == "" {
		if env := os.Getenv("KUBECONFIG"); env != "" {
			kubeconfig = env
		} else if home, err := os.UserHomeDir(); err == nil {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	// Prefer kubeconfig file if available
	if cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig); err == nil {
		return cfg
	}

	// Fallback to in-cluster config
	if cfg, err := rest.InClusterConfig(); err == nil {
		return cfg
	}

	utils.Panic("Failed to get kubeconfig")
	return nil
}

func GetKubeClient(ctx context.Context) *clientset.Clientset {
	once.Do(func() {
		kubeClient = clientset.NewForConfigOrDie(GetKubeConfig())
		log.Info(ctx, "Get Kubernetes Client OK")
	})
	return kubeClient
}
