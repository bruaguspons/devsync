package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bruaguspons/devsync/core"
	"github.com/bruaguspons/devsync/providers/mise"
	"github.com/bruaguspons/devsync/providers/skills"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print devsync version and exit")
	yes := flag.Bool("yes", false, "apply every pending change without an interactive prompt (for non-TTY contexts, e.g. a Docker build)")
	flag.Parse()

	if *showVersion {
		fmt.Println("devsync " + version)
		return
	}

	providers := []core.Provider{
		mise.New(mise.DefaultSnapshotPath()),
		skills.New(skills.DefaultDesiredRoot(), skills.DefaultInstalledRoot()),
	}

	if err := core.Run(providers, core.DefaultLockPath(), *yes); err != nil {
		fmt.Fprintln(os.Stderr, "devsync: "+err.Error())
		os.Exit(1)
	}
}
