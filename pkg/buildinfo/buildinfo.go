// Package buildinfo provides build information such as version, revision, branch, build user and build date.
package buildinfo

import (
	"github.com/prometheus/common/version"
)

var (
	Version   = "snapshot"
	Revision  = "unknown"
	Branch    = "unknown"
	BuildDate = "unknown"
)

func init() {
	version.Version = Version
	version.Revision = Revision
	version.Branch = Branch
	version.BuildDate = BuildDate
}
