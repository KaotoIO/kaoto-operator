package defaults

import "time"

const (
	SyncInterval       = 5 * time.Second
	RetryInterval      = 10 * time.Second
	ConflictInterval   = 1 * time.Second
	KaotoFinalizerName = "kaoto.io/finalizer"
)

var (
	KaotoAppImage = "quay.io/kaotoio/kaoto-app:main"
)
