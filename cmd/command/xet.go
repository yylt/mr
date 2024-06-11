package command

import (
	"github.com/spf13/cobra"
)

func init() {
	setCmd.AddCommand(cntCmd)
}

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "提供系统源，语言和容器镜像",
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "查询支持的系统源",
}
