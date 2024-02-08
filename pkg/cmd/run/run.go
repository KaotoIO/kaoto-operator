package run

import (
	"fmt"
	"runtime"

	"github.com/kaotoIO/kaoto-operator/internal/controller/designer"
	"github.com/kaotoIO/kaoto-operator/pkg/controller"
	"github.com/kaotoIO/kaoto-operator/pkg/defaults"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	rtcache "sigs.k8s.io/controller-runtime/pkg/cache"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	designerApi "github.com/kaotoIO/kaoto-operator/api/designer/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

func init() {
	utilruntime.Must(designerApi.AddToScheme(controller.Scheme))
	utilruntime.Must(routev1.Install(controller.Scheme))
}

func NewRunCmd() *cobra.Command {

	options := controller.Options{
		MetricsAddr:                   ":8080",
		ProbeAddr:                     ":8081",
		PprofAddr:                     "",
		LeaderElectionID:              "9aa9f118.kaoto.io",
		EnableLeaderElection:          true,
		ReleaseLeaderElectionOnCancel: true,
		LeaderElectionNamespace:       "",
	}

	cmd := cobra.Command{
		Use:   "run",
		Short: "run",
		RunE: func(cmd *cobra.Command, args []string) error {
			return controller.Start(options, func(manager manager.Manager, opts controller.Options) error {
				l := ctrl.Log.WithName("run")
				l.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
				l.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
				l.Info(fmt.Sprintf("Kaoto App Image: %s", defaults.KaotoAppImage))

				selector, err := designer.AppSelector()
				if err != nil {
					return errors.Wrap(err, "unable to compute cache's watch selector")
				}

				options.WatchSelectors = map[rtclient.Object]rtcache.ByObject{
					&appsv1.Deployment{}:         {Label: selector},
					&netv1.Ingress{}:             {Label: selector},
					&routev1.Route{}:             {Label: selector},
					&corev1.Secret{}:             {Label: selector},
					&rbacv1.ClusterRoleBinding{}: {Label: selector},
					&corev1.ServiceAccount{}:     {Label: selector},
				}

				rec, err := designer.NewKaotoReconciler(manager)
				if err != nil {
					return err
				}

				return rec.SetupWithManager(cmd.Context(), manager)
			})
		},
	}

	cmd.Flags().StringVar(&options.LeaderElectionID, "leader-election-id", options.LeaderElectionID, "The leader election ID of the operator.")
	cmd.Flags().StringVar(&options.LeaderElectionNamespace, "leader-election-namespace", options.LeaderElectionNamespace, "The leader election namespace.")
	cmd.Flags().BoolVar(&options.EnableLeaderElection, "leader-election", options.EnableLeaderElection, "Enable leader election for controller manager.")
	cmd.Flags().BoolVar(&options.ReleaseLeaderElectionOnCancel, "leader-election-release", options.ReleaseLeaderElectionOnCancel, "If the leader should step down voluntarily.")

	cmd.Flags().StringVar(&options.MetricsAddr, "metrics-bind-address", options.MetricsAddr, "The address the metric endpoint binds to.")
	cmd.Flags().StringVar(&options.ProbeAddr, "health-probe-bind-address", options.ProbeAddr, "The address the probe endpoint binds to.")
	cmd.Flags().StringVar(&options.PprofAddr, "pprof-bind-address", options.PprofAddr, "The address the pprof endpoint binds to.")

	return &cmd
}
