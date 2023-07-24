package defaults

import "time"

const (
	SyncInterval       = 5 * time.Second
	RetryInterval      = 10 * time.Second
	ConflictInterval   = 1 * time.Second
	KaotoFinalizerName = "kaoto.io/finalizer"
)

var (
	KaotoStandaloneImage = "quay.io/kaotoio/standalone:main-jvm"
)
