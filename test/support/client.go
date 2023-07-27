package support

import (
	"errors"
	"os"
	"path/filepath"

	designerApi "github.com/kaotoIO/kaoto-operator/apis/designer/v1alpha1"
	"github.com/kaotoIO/kaoto-operator/pkg/client"
	routev1 "github.com/openshift/api/route/v1"
	route "github.com/openshift/client-go/route/clientset/versioned"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Client struct {
	kubernetes.Interface

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

	c := Client{
		scheme:    runtime.NewScheme(),
		Interface: kubeClient,
		Discovery: discoveryClient,
		config:    cfg,
	}

	if err := designerApi.AddToScheme(c.scheme); err != nil {
		return nil, err
	}
	if err := routev1.Install(c.scheme); err != nil {
		return nil, err
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
