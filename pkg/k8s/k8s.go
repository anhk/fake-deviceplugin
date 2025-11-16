package k8s

import (
	"fake-deviceplugin/pkg/utils"
	"sync"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeconfig string = ""
var kubeClient *clientset.Clientset
var once sync.Once

func GetKubeConfig() *rest.Config {
	// Prefer kubeconfig file

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

func GetKubeClient() *clientset.Clientset {
	once.Do(func() {
		kubeClient = clientset.NewForConfigOrDie(GetKubeConfig())
	})
	return kubeClient
}
