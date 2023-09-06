package controller

import (
	rtcache "sigs.k8s.io/controller-runtime/pkg/cache"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Options struct {
	MetricsAddr                   string
	ProbeAddr                     string
	PprofAddr                     string
	LeaderElectionID              string
	LeaderElectionNamespace       string
	EnableLeaderElection          bool
	ReleaseLeaderElectionOnCancel bool
	WatchSelectors                map[rtclient.Object]rtcache.ByObject
}
