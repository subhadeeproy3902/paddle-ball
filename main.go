package main

import "github.com/subhadeeproy3902/paddle-ball/cmd"

// Injected by goreleaser at build time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.Execute(version, commit, date)
}
