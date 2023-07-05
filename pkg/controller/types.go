package controller

type Options struct {
	MetricsAddr                   string
	ProbeAddr                     string
	ProofAddr                     string
	LeaderElectionID              string
	LeaderElectionNamespace       string
	EnableLeaderElection          bool
	ReleaseLeaderElectionOnCancel bool
}
