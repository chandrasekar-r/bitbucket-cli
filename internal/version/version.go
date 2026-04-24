package version

// These variables are injected at build time via -ldflags.
// Credentials are NEVER stored in source — they are injected from
// GitHub Actions secrets at release time via GoReleaser ldflags.
//
//	-X github.com/chandrasekar-r/bitbucket-cli/internal/version.Version=v1.0.0
//	-X github.com/chandrasekar-r/bitbucket-cli/internal/version.BuildDate=2026-01-01
//	-X github.com/chandrasekar-r/bitbucket-cli/internal/version.Commit=abc1234
//	-X github.com/chandrasekar-r/bitbucket-cli/internal/version.OAuthClientID=<key>
//	-X github.com/chandrasekar-r/bitbucket-cli/internal/version.OAuthClientSecret=<secret>
var (
	Version           = "dev"
	BuildDate         = "unknown"
	Commit            = "none"
	OAuthClientID     = ""
	OAuthClientSecret = ""
)
