package command

import (
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(setCmd, getCmd)
}

var RootCmd = &cobra.Command{
	Use:   "mr",
	Short: "mirror tool for repo, language, software(docker, container), etc...",
}
