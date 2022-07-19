package controllers

import (
	kaotoiov1alpha1 "github.com/kaotoIO/kaoto-operator/api/v1alpha1"
	v12 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateRoleBinding(role *v12.Role) *v12.RoleBinding {
	binding := v12.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "integrator-role-binding",
			Namespace: role.Namespace,
		},
		RoleRef: v12.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Name:     role.Name,
			Kind:     "Role",
		},
		Subjects: []v12.Subject{{
			Kind:      v12.ServiceAccountKind,
			Name:      "default",
			Namespace: role.Namespace,
		}},
	}
	return &binding
}

func CreateClusterRoleBinding(role *v12.ClusterRole, namespace string) *v12.ClusterRoleBinding {
	binding := v12.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "integrator-role-binding-cr",
			Namespace: role.Namespace,
		},
		RoleRef: v12.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Name:     role.Name,
			Kind:     "ClusterRole",
		},
		Subjects: []v12.Subject{{
			Kind:      v12.ServiceAccountKind,
			Name:      "default",
			Namespace: namespace,
		}},
	}
	return &binding
}

func CreateIntegratorClusterRole(kaoto kaotoiov1alpha1.Kaoto) *v12.ClusterRole {
	role := &v12.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "integrator-role-cr",
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
