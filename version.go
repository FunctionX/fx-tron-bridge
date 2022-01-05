package fxtronbridge

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var Version string
var Commit string
var BuildTime string

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "Show version",
		Run: func(*cobra.Command, []string) {
			fmt.Printf("Version: %s\n", Version)
			fmt.Printf("Git Commit: %s\n", Commit)
			fmt.Printf("Go Version: %s\n", runtime.Version())
			fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
			fmt.Printf("Build Time: %s\n", BuildTime)
		},
	}
}
