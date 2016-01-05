package global

import (
	"github.com/wakeful-deployment/operator/fsm"
)

type GlobalInfo struct {
	Nodename   string
	Consulhost string
}

type GlobalMetadata map[string]string

var (
	Initial                      = fsm.State{Name: "Initial"}
	ConfigFailed                 = fsm.State{Name: "ConfigFailed"}
	Booting                      = fsm.State{Name: "Booting"}
	PostingMetadataFailed        = fsm.State{Name: "PostingMetadataFailed"}
	ReattemptingToBoot           = fsm.State{Name: "ReattemptingToBoot"}
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
	ReattemptingToBoot,
	Booted,
	FetchingNodeStateFailed,
	MergingStateFailed,
	NormalizingFailed,
	FetchingDirectoryStateFailed,
	AttemptingToRecover,
	Running,
}

var Failures = []fsm.State{ConsulFailed, FetchingNodeStateFailed, NormalizingFailed}

var AllowedTransitions = fsm.Rules{
	fsm.From(Initial).To(Booting, ConfigFailed),
	fsm.From(ConfigFailed).To(),
	fsm.From(Booting).To(append(Failures, Booted)...),
	fsm.From(PostingMetadataFailed).To(ReattemptingToBoot),
	fsm.From(ConsulFailed).To(ReattemptingToBoot, AttemptingToRecover),
	fsm.From(ReattemptingToBoot).To(append(Failures, Booted)...),
	fsm.From(Booted).To(append(Failures, Running)...),
	fsm.From(FetchingNodeStateFailed).To(ReattemptingToBoot, AttemptingToRecover),
	fsm.From(MergingStateFailed).To(ReattemptingToBoot, AttemptingToRecover),
	fsm.From(NormalizingFailed).To(ReattemptingToBoot, AttemptingToRecover),
	fsm.From(FetchingDirectoryStateFailed).To(ReattemptingToBoot, AttemptingToRecover),
	fsm.From(AttemptingToRecover).To(append(Failures, Running)...),
	fsm.From(Running).To(append(Failures, Running)...),
}

var Machine = fsm.Machine{CurrentState: Initial, Rules: AllowedTransitions, States: states}

var Info = GlobalInfo{}
var Metadata = GlobalMetadata{}
