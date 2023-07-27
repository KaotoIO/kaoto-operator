package support

import (
	"context"
	"sync"
	"testing"

	"github.com/onsi/gomega"
)

type Test interface {
	T() *testing.T
	Ctx() context.Context
	Client() *Client

	gomega.Gomega
}

type Option[T any] interface {
	applyTo(to T) error
}

type errorOption[T any] func(to T) error

// nolint: unused
// To be removed when the false-positivity is fixed.
func (o errorOption[T]) applyTo(to T) error {
	return o(to)
}

var _ Option[any] = errorOption[any](nil)

func With(t *testing.T) Test {
	t.Helper()
	ctx := context.Background()
	if deadline, ok := t.Deadline(); ok {
		withDeadline, cancel := context.WithDeadline(ctx, deadline)
		t.Cleanup(cancel)
		ctx = withDeadline
	}

	return &T{
		WithT: gomega.NewWithT(t),
		t:     t,
		ctx:   ctx,
	}
}

type T struct {
	*gomega.WithT

	t      *testing.T
	client *Client
	once   sync.Once

	// nolint: containedctx
	ctx context.Context
}

func (t *T) T() *testing.T {
	return t.t
}

func (t *T) Ctx() context.Context {
	return t.ctx
}

func (t *T) Client() *Client {
	t.once.Do(func() {
		c, err := newClient()
		if err != nil {
			t.T().Fatalf("Error creating client: %v", err)
		}
		t.client = c
	})
	return t.client
}
