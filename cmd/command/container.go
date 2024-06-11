package command

import (
	"log"
	"os/user"
	"runtime"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/yylt/mr/pkg/image"
	"github.com/yylt/mr/pkg/util"
)

var (
	regitry   string = "registry"
	regmirror string = "mirror"
)

func init() {
	cntCmd.Flags().StringP(regitry, "r", "", "设置仓库, 常见有 docker.io, gcr.io, ghcr.io 等")
	cntCmd.Flags().StringArrayP(regmirror, "m", nil, "设置仓库的镜像")
	//cntCmd.MarkFlagRequired(regitry)
	cntCmd.MarkFlagRequired(regmirror)
}

var cntCmd = &cobra.Command{
	Use:     "container",
	Aliases: []string{"cnt"},
	Short:   "设置容器镜像的mirror, 支持 docker 和 container. 优先检测 dockerd 进程",

	RunE: func(cmd *cobra.Command, args []string) error {
		mustValidEnv()
		r, err := cmd.Flags().GetString(regitry)
		if err != nil {
			log.Fatalf("选项 '%s' 失败: %v", regitry, err)
		}
		if r == "" {
			r = "docker.io"
		}
		ms, err := cmd.Flags().GetStringArray(regmirror)
		if err != nil {
			log.Fatalf("选项 '%s' 失败: %v", regmirror, err)
		}
		return image.NewImg(r, ms).SetMirror()
	},
}

func mustValidEnv() {
	if !util.IsSupportOs(string(util.Linux)) {
		log.Panic("不支持的OS: ", runtime.GOOS)
	}
	currentUser, err := user.Current()
	if err != nil {
		log.Panic("用户信息失败: ", err)
	}

	if currentUser.Uid != "0" && syscall.Geteuid() != 0 {
		log.Panic("请以 root 或 sudo 启动")
	}
}
