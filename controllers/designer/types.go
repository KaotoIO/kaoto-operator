package designer

import (
	"context"

	"github.com/kaotoIO/kaoto-operator/pkg/controller/client"

	"sigs.k8s.io/controller-runtime/pkg/builder"

	"k8s.io/apimachinery/pkg/api/resource"

	kaotoiov1alpha1 "github.com/kaotoIO/kaoto-operator/apis/designer/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

type ClusterType string

var (
	KaotoStandaloneDefaultMemory = resource.MustParse("600Mi")
	KaotoStandaloneDefaultCPU    = resource.MustParse("500m")
)

const (
	ClusterTypeVanilla   ClusterType = "Vanilla"
	ClusterTypeOpenShift ClusterType = "OpenShift"

	K

	KaotoAppName              string = "kaoto"
	KaotoComponentDesigner    string = "designer"
	KaotoOperatorFieldManager string = "kaoto-operator"
	KaotoPort                 int32  = 8080
	KaotoPortType             string = "http"
	KaotoLivenessProbePath    string = "/"
	KaotoReadinessProbePath   string = "/"

	KubernetesLabelAppName      = "app.kubernetes.io/name"
	KubernetesLabelAppInstance  = "app.kubernetes.io/instance"
	KubernetesLabelAppComponent = "app.kubernetes.io/component"
	KubernetesLabelAppPartOf    = "app.kubernetes.io/part-of"
	KubernetesLabelAppManagedBy = "app.kubernetes.io/managed-by"
)

type ReconciliationRequest struct {
	*client.Client
	types.NamespacedName

	ClusterType ClusterType
	Kaoto       *kaotoiov1alpha1.Kaoto
}

type Action interface {
	Configure(context.Context, *client.Client, *builder.Builder) (*builder.Builder, error)
	Apply(context.Context, *ReconciliationRequest) error
	Cleanup(context.Context, *ReconciliationRequest) error
}
