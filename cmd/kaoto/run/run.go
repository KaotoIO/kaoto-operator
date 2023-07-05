package run

import (
	"github.com/kaotoIO/kaoto-operator/controllers"
	"github.com/kaotoIO/kaoto-operator/pkg/controller"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	kaotoiov1alpha1 "github.com/kaotoIO/kaoto-operator/api/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

func init() {
	utilruntime.Must(kaotoiov1alpha1.AddToScheme(controller.Scheme))
	utilruntime.Must(routev1.AddToScheme(controller.Scheme))
}

func NewRunCmd() *cobra.Command {
	options := controller.Options{
		MetricsAddr:                   ":8080",
		ProbeAddr:                     ":8081",
		ProofAddr:                     "",
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
				rec := controllers.KaotoReconciler{
					Client: manager.GetClient(),
					Scheme: manager.GetScheme(),
				}

				return rec.SetupWithManager(manager)
			})
		},
	}

	cmd.Flags().StringVar(&options.LeaderElectionID, "leader-election-id", options.LeaderElectionID, "The leader election ID of the operator.")
	cmd.Flags().StringVar(&options.LeaderElectionNamespace, "leader-election-namespace", options.LeaderElectionNamespace, "The leader election namespace.")
	cmd.Flags().StringVar(&options.MetricsAddr, "metrics-bind-address", options.MetricsAddr, "The address the metric endpoint binds to.")
	cmd.Flags().StringVar(&options.ProbeAddr, "health-probe-bind-address", options.ProbeAddr, "The address the probe endpoint binds to.")
	cmd.Flags().StringVar(&options.ProofAddr, "pprof-bind-address", options.ProofAddr, "The address the pprof endpoint binds to.")
	cmd.Flags().BoolVar(&options.EnableLeaderElection, "leader-election", options.EnableLeaderElection, "Enable leader election for controller manager.")
	cmd.Flags().BoolVar(&options.ReleaseLeaderElectionOnCancel, "leader-election-release", options.ReleaseLeaderElectionOnCancel, "If the leader should step down voluntarily.")

	_ = cmd.MarkFlagRequired("operator-id")
	_ = cmd.MarkFlagRequired("operator-group")
	_ = cmd.MarkFlagRequired("operator-type")
	_ = cmd.MarkFlagRequired("operator-version")

	return &cmd
}
