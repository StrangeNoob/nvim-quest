package main

import "github.com/StrangeNoob/nvim-quest/cmd"

// Injected at build time via -ldflags (see .goreleaser.yaml). Defaults apply to
// `go run` / `go build` dev builds.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.Execute(version, commit, date)
}
