package controllers

import (
	kaotoiov1alpha1 "github.com/kaotoIO/kaoto-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateRoleBinding(role *v12.Role, account *v1.ServiceAccount) *v12.RoleBinding {
	binding := v12.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "integrator-role-binding",
			Namespace: role.Namespace,
		},
		RoleRef: v12.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     role.Kind,
			Name:     role.Name,
		},
		Subjects: []v12.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      account.Name,
				Namespace: account.Namespace,
			},
		},
	}
	return &binding
}

func CreateIntegratorRole(kaoto kaotoiov1alpha1.Kaoto) *v12.Role {
	role := &v12.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "integrator-role",
			Namespace: kaoto.Namespace,
		},
		Rules: []v12.PolicyRule{{
			Verbs:     []string{"create", "delete", "update", "watch", "get", "list", "patch"},
			APIGroups: []string{"camel.apache.org"},
			Resources: []string{"kamelets", "kameletbindings"},
		},
		},
	}
	return role
}
