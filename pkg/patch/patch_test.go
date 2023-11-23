package patch

import (
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMergePatch(t *testing.T) {
	d1 := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
	}
	d2 := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
	}

	data, err := MergePatch(d1, d2)

	require.NoError(t, err)
	require.NotNil(t, data)
	require.Empty(t, data, 0)
}
