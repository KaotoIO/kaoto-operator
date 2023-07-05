package logger

import "sigs.k8s.io/controller-runtime/pkg/log/zap"

var (
	Options = zap.Options{
		Development: true,
	}
)
