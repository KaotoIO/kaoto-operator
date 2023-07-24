package designer

import (
	"context"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/kaotoIO/kaoto-operator/config/client"

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
	KaotoStandaloneName            string = "kaoto-standalone"
	KaotoOperatorFieldManager      string = "kaoto-operator"
	KaotoDeploymentClusterRoleName string = "kaoto-backend"
	KaotoPort                      int32  = 8081
	KaotoPortType                  string = "http"
	KaotoLivenessProbePath         string = "/q/health/live"
	KaotoReadinessProbePath        string = "/q/health/ready"
)

type ReconciliationRequest struct {
	*client.Client
	types.NamespacedName

	ClusterType ClusterType
	Kaoto       *kaotoiov1alpha1.Kaoto
}

type Action interface {
	Apply(ctx context.Context, rr *ReconciliationRequest) error
}
