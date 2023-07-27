package e2e

import (
	"testing"

	. "github.com/kaotoIO/kaoto-operator/test/support"
)

func TestDesignerDeploy(t *testing.T) {
	test := With(t)
	test.T().Parallel()
}
