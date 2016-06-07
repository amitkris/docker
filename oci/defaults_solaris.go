package oci

import (
	"runtime"

	"github.com/opencontainers/specs/specs-go"
)

// DefaultSpec returns default oci spec used by docker.
func DefaultSpec() specs.Spec {
	s := specs.Spec{
		Version: "0.6.0",
		Platform: specs.Platform{
			OS:   "SunOS",
			Arch: runtime.GOARCH,
		},
	}
	return s
}
