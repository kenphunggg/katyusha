package fukabunsan // 負荷分散 - ふかぶんさん - Load Balancing

import (
	"log"
	"os"
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes" // KUBECONFIG
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func KubeConfig() *rest.Config {
	var config *rest.Config
	var err error

	// Check if a kubeconfig file exists in the default location.
	home := homedir.HomeDir()
	if home != "" {
		kubeconfigPath := filepath.Join(home, ".kube", "config")
		if _, err = os.Stat(kubeconfigPath); err == nil {
			// File exists, use it.
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
			if err != nil {
				log.Fatalf("Failed to build config from kubeconfig: %v", err)
			}
			return config
		}
	}

	// Fallback to in-cluster config
	config, err = rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to load in-cluster config: %v", err)
	}

	return config
}

func GetClientSet() *kubernetes.Clientset {
	clientset, err := kubernetes.NewForConfig(KUBECONFIG)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}

func GetDynamicClient() *dynamic.DynamicClient {
	dynclient, err := dynamic.NewForConfig(KUBECONFIG)
	if err != nil {
		panic(err.Error())
	}
	return dynclient
}
