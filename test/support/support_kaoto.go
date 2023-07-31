package support

import (
	kaoto "github.com/kaotoIO/kaoto-operator/apis/designer/v1alpha1"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Kaoto(t Test, namespace *corev1.Namespace, name string) func(g gomega.Gomega) *kaoto.Kaoto {
	return func(g gomega.Gomega) *kaoto.Kaoto {
		answer, err := t.Client().Kaoto.DesignerV1alpha1().Kaotos(namespace.Name).Get(
			t.Ctx(),
			name,
			metav1.GetOptions{},
		)

		g.Expect(err).NotTo(gomega.HaveOccurred())

		return answer
	}
}
