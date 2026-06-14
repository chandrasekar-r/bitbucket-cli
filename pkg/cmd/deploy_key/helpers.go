package deploykeycmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/gitcontext"
)

func repoContext(_ *cmdutil.Factory) (workspace, slug string, err error) {
	if ctx := gitcontext.FromRemote(); ctx != nil {
		return ctx.Workspace, ctx.RepoSlug, nil
	}
	return "", "", fmt.Errorf("no repo found: run from inside a cloned Bitbucket repository")
}

func readKeyMaterial(key, keyFile string) (string, error) {
	switch {
	case key != "" && keyFile != "":
		return "", fmt.Errorf("only one of --key or --key-file may be set")
	case key != "":
		return strings.TrimSpace(key), nil
	case keyFile != "":
		data, err := os.ReadFile(keyFile)
		if err != nil {
			return "", fmt.Errorf("reading key file: %w", err)
		}
		return strings.TrimSpace(string(data)), nil
	default:
		return "", fmt.Errorf("one of --key or --key-file is required")
	}
}