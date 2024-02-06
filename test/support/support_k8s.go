package support

import (
	kaotoApi "github.com/kaotoIO/kaoto-operator/api/designer/v1alpha1"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Deployment(t Test, kaoto *kaotoApi.Kaoto) func(g gomega.Gomega) (*appsv1.Deployment, error) {
	return func(g gomega.Gomega) (*appsv1.Deployment, error) {
		answer, err := t.Client().AppsV1().Deployments(kaoto.Namespace).Get(
			t.Ctx(),
			kaoto.Name,
			metav1.GetOptions{},
		)

		if k8serrors.IsNotFound(err) {
			return nil, nil
		}

		return answer, err
	}
}

func Service(t Test, kaoto *kaotoApi.Kaoto) func(g gomega.Gomega) (*corev1.Service, error) {
	return func(g gomega.Gomega) (*corev1.Service, error) {
		answer, err := t.Client().CoreV1().Services(kaoto.Namespace).Get(
			t.Ctx(),
			kaoto.Name,
			metav1.GetOptions{},
		)

		if k8serrors.IsNotFound(err) {
			return nil, nil
		}

		return answer, err
	}
}

func ServiceAccount(t Test, kaoto *kaotoApi.Kaoto) func(g gomega.Gomega) (*corev1.ServiceAccount, error) {
	return func(g gomega.Gomega) (*corev1.ServiceAccount, error) {
		answer, err := t.Client().CoreV1().ServiceAccounts(kaoto.Namespace).Get(
			t.Ctx(),
			kaoto.Name,
			metav1.GetOptions{},
		)

		if k8serrors.IsNotFound(err) {
			return nil, nil
		}

		return answer, err
	}
}

func ClusterRoleBinding(t Test, kaoto *kaotoApi.Kaoto) func(g gomega.Gomega) (*rbacv1.ClusterRoleBinding, error) {
	return func(g gomega.Gomega) (*rbacv1.ClusterRoleBinding, error) {
		answer, err := t.Client().RbacV1().ClusterRoleBindings().Get(
			t.Ctx(),
			kaoto.Namespace+"-"+kaoto.Name,
			metav1.GetOptions{},
		)

		if k8serrors.IsNotFound(err) {
			return nil, nil
		}

		return answer, err
	}
}

func GetIngress(t Test, kaoto *kaotoApi.Kaoto) (*netv1.Ingress, error) {
	answer, err := t.Client().NetworkingV1().Ingresses(kaoto.Namespace).Get(
		t.Ctx(),
		kaoto.Name,
		metav1.GetOptions{},
	)

	if k8serrors.IsNotFound(err) {
		return nil, nil
	}

	return answer, err
}

func Ingress(t Test, kaoto *kaotoApi.Kaoto) func(g gomega.Gomega) (*netv1.Ingress, error) {
	return func(g gomega.Gomega) (*netv1.Ingress, error) {
		return GetIngress(t, kaoto)
	}
}
