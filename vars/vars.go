package vars

// These variables are set at build time (using -ldflags -X ...)
// See Taskfile.yaml for more details.

var (
	ClerkRootURL = "unset"
	LoginURL     = "unset"
	Stage        = "unset"
	CommitID     = "unset"
	Version      = "unset"
	BuildDate    = "unset"
	Branch       = "unset"
	GCSBucket    = "unset"
	GCSKey       = "unset"
)
