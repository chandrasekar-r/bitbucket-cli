package version

// These variables are injected at build time via -ldflags:
//
//	-X github.com/chandrasekar-r/bitbucket-cli/internal/version.Version=v1.0.0
//	-X github.com/chandrasekar-r/bitbucket-cli/internal/version.BuildDate=2026-01-01
//	-X github.com/chandrasekar-r/bitbucket-cli/internal/version.Commit=abc1234
//	-X github.com/chandrasekar-r/bitbucket-cli/internal/version.OAuthClientID=<id>
var (
	Version       = "dev"
	BuildDate     = "unknown"
	Commit        = "none"
	OAuthClientID = ""
)
