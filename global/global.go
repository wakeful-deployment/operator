package global

import (
	"github.com/wakeful-deployment/operator/fsm"
)

var (
	Initial                      = fsm.State{Name: "Initial"}
	ConfigFailed                 = fsm.State{Name: "ConfigFailed"}
	Booting                      = fsm.State{Name: "Booting"}
	PostingMetadataFailed        = fsm.State{Name: "PostingMetadataFailed"}
	Booted                       = fsm.State{Name: "Booted"}
	ConsulFailed                 = fsm.State{Name: "ConsulFailed"}
	FetchingNodeStateFailed      = fsm.State{Name: "FetchingNodeStateFailed"}
	MergingStateFailed           = fsm.State{Name: "MergingStateFailed"}
	NormalizingFailed            = fsm.State{Name: "NormalizingFailed"}
	FetchingDirectoryStateFailed = fsm.State{Name: "FetchingDirectoryStateFailed"}
	AttemptingToRecover          = fsm.State{Name: "AttemptingToRecover"}
	Running                      = fsm.State{Name: "Running"}
)

var states = []fsm.State{
	Initial,
	ConfigFailed,
	Booting,
	PostingMetadataFailed,
	ConsulFailed,
	Booted,
	FetchingNodeStateFailed,
	MergingStateFailed,
	NormalizingFailed,
	FetchingDirectoryStateFailed,
	AttemptingToRecover,
	Running,
}

var AllowedTransitions = fsm.Rules{
	fsm.From(Initial).To(Booting, ConfigFailed),
	fsm.From(ConfigFailed).To(),
	fsm.From(Booting).To(ConsulFailed, PostingMetadataFailed, Booted),
	fsm.From(PostingMetadataFailed).To(Booting),
	fsm.From(ConsulFailed).To(Booting, AttemptingToRecover),
	fsm.From(Booted).To(ConsulFailed, FetchingDirectoryStateFailed, FetchingNodeStateFailed, NormalizingFailed, Running),
	fsm.From(FetchingNodeStateFailed).To(AttemptingToRecover),
	fsm.From(MergingStateFailed).To(AttemptingToRecover),
	fsm.From(NormalizingFailed).To(AttemptingToRecover),
	fsm.From(FetchingDirectoryStateFailed).To(AttemptingToRecover),
	fsm.From(AttemptingToRecover).To(ConsulFailed, FetchingNodeStateFailed, NormalizingFailed, Running),
	fsm.From(Running).To(ConsulFailed, FetchingNodeStateFailed, NormalizingFailed, Running),
}

var Machine = fsm.Machine{CurrentState: Initial, Rules: AllowedTransitions, States: states}
