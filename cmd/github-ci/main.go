package main

import (
	"github.com/reugn/github-ci/internal/cmd"
)

// version is set by goreleaser via ldflags
var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
