package support

import (
	"github.com/onsi/gomega"
	"github.com/rs/xid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createTestNamespace(t Test, options ...Option[*corev1.Namespace]) *corev1.Namespace {
	t.T().Helper()

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kaoto-test-e2e-" + xid.New().String(),
		},
	}

	for _, option := range options {
		t.Expect(option.applyTo(namespace)).To(gomega.Succeed())
	}

	namespace, err := t.Client().CoreV1().Namespaces().Create(t.Ctx(), namespace, metav1.CreateOptions{})

	t.Expect(err).NotTo(gomega.HaveOccurred())

	return namespace
}

func deleteTestNamespace(t Test, namespace *corev1.Namespace) {
	t.T().Helper()
	propagationPolicy := metav1.DeletePropagationBackground
	err := t.Client().CoreV1().Namespaces().Delete(t.Ctx(), namespace.Name, metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
	t.Expect(err).NotTo(gomega.HaveOccurred())
}
