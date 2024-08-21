package e2e

import (
	"testing"

	"github.com/kaotoIO/kaoto-operator/pkg/defaults"

	corev1 "k8s.io/api/core/v1"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/rs/xid"

	kaoto "github.com/kaotoIO/kaoto-operator/api/designer/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/kaotoIO/kaoto-operator/test/support"
	. "github.com/onsi/gomega"
)

func TestDesignerDeploy(t *testing.T) {
	t.Parallel()

	tests := map[string]kaoto.Kaoto{
		"no-ingress": {},
		"with-ingress": {
			Spec: kaoto.KaotoSpec{
				Ingress: &kaoto.IngressSpec{},
			},
		},
	}

	for name, res := range tests {
		instance := res

		t.Run(name, func(t *testing.T) {
			test := With(t)
			test.T().Parallel()

			ns := test.NewTestNamespace()

			instance.Name = xid.New().String()
			instance.Namespace = ns.Name

			kd, err := test.Client().Kaoto.DesignerV1alpha1().Kaotoes(ns.Name).Create(
				test.Ctx(),
				&instance,
				metav1.CreateOptions{})

			test.T().Cleanup(func() {
				err := test.Client().Kaoto.DesignerV1alpha1().Kaotoes(ns.Name).Delete(
					test.Ctx(),
					kd.Name,
					metav1.DeleteOptions{},
				)

				test.Expect(err).NotTo(HaveOccurred())
			})

			test.Expect(err).NotTo(HaveOccurred())

			test.Eventually(Service(test, kd), TestTimeoutLong).Should(
				Not(BeNil()))
			test.Eventually(Deployment(test, kd), TestTimeoutLong).Should(
				WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))
			test.Eventually(Deployment(test, kd), TestTimeoutLong).Should(
				WithTransform(ContainerImage(0), Equal(defaults.KaotoAppImage)))

			if kd.Spec.Ingress == nil {
				test.Expect(GetIngress(test, kd)).Should(
					BeNil())
			} else {
				test.Expect(GetIngress(test, kd)).Should(
					Not(BeNil()))
			}
		})
	}
}
