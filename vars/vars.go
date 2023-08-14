package vars

// These variables are set at build time (using -ldflags -X ...)

var (
	ClerkRootURL = "https://welcomed-snapper-45.clerk.accounts.dev/"
	LoginURL     = "https://ampersand-cli-auth-dev.web.app"
	Stage        = "dev"
	CommitID     = "unknown"
	Version      = "latest"
	BuildDate    = "unknown"
	Branch       = "unknown"
)
