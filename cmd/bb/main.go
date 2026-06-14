package main

import (
	"os"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/alias"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmd/root"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/config"
)

func main() {
	args := os.Args[1:]
	if cfg, err := config.Load(); err == nil {
		args = alias.Expand(args, cfg)
	}
	root.ExecuteWithArgs(args)
}