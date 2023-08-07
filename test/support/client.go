package support

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/kaotoIO/kaoto-operator/pkg/controller/client"

	route "github.com/openshift/client-go/route/clientset/versioned"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	kaotoClient "github.com/kaotoIO/kaoto-operator/pkg/client/kaoto/clientset/versioned"
)

type Client struct {
	kubernetes.Interface

	Kaoto     kaotoClient.Interface
	Discovery discovery.DiscoveryInterface
	Route     route.Interface

	scheme *runtime.Scheme
	config *rest.Config
}

func newClient() (*Client, error) {
	kc := os.Getenv("KUBECONFIG")
	if kc == "" {
		home := homedir.HomeDir()
		if home != "" {
			kc = filepath.Join(home, ".kube", "config")
		}
	}

	if kc == "" {
		return nil, errors.New("unable to determine KUBECONFIG")
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", kc)
	if err != nil {
		return nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	kClient, err := kaotoClient.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	c := Client{
		Interface: kubeClient,
		Discovery: discoveryClient,
		Kaoto:     kClient,
		config:    cfg,
	}

	io, err := client.IsOpenShift(discoveryClient)
	if err != nil {
		return nil, err
	}

	if io {
		routeClient, err := route.NewForConfig(cfg)
		if err != nil {
			return nil, err
		}

		c.Route = routeClient
	}

	return &c, nil
}
