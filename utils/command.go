package utils

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// CheckErr prints the msg with the prefix 'Error:' and exits with error code 1. If the msg is nil, it does nothing.
func CheckErr(err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "\033[1;31m%s\033[0m", fmt.Sprintf("Error: %s\n", err.Error()))
		os.Exit(1)
	}
}

func SilenceCmdErrors(cmd *cobra.Command) {
	cmd.SilenceErrors = true
	for _, subCmd := range cmd.Commands() {
		SilenceCmdErrors(subCmd)
	}
}

func AddFlags(cmd *cobra.Command, name string, value interface{}, usage string, required bool) {
	switch v := value.(type) {
	case string:
		cmd.Flags().String(name, v, usage)
	case int64:
		cmd.Flags().Int64(name, v, usage)
	case uint64:
		cmd.Flags().Uint64(name, v, usage)
	case int:
		cmd.Flags().Int(name, v, usage)
	case uint:
		cmd.Flags().Uint(name, v, usage)
	case float64:
		cmd.Flags().Float64(name, v, usage)
	case float32:
		cmd.Flags().Float32(name, v, usage)
	case bool:
		cmd.Flags().Bool(name, v, usage)
	default:
		panic("Invalid flag type")
	}
	if required {
		CheckErr(cmd.MarkFlagRequired(name))
	}
}
