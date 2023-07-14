package designer

import (
	"context"

	kaotoiov1alpha1 "github.com/kaotoIO/kaoto-operator/apis/designer/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterType string

const (
	ClusterTypeVanilla   ClusterType = "Vanilla"
	ClusterTypeOpenShift ClusterType = "OpenShift"
)

type ReconciliationRequest struct {
	client.Client
	types.NamespacedName

	ClusterType ClusterType
	Kaoto       *kaotoiov1alpha1.Kaoto
}

type Action interface {
	Apply(ctx context.Context, rr *ReconciliationRequest) error
}
