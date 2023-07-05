package patch

import (
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
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

	assert.Nil(t, err)
	assert.NotNil(t, data)
	assert.Len(t, data, 0)
}
