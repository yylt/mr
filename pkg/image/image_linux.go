package image

import (
	"encoding/json"
	"io/fs"
	"log"
	"os"
	"strings"
	"syscall"

	ccommand "github.com/containerd/containerd/cmd/containerd/command"
	criconfig "github.com/containerd/containerd/pkg/cri/config"

	"github.com/pelletier/go-toml"
	"github.com/prometheus/procfs"
	"github.com/yylt/mr/pkg/util"
)

var (
	dockerproc     = "dockerd"
	containerdproc = "containerd"

	fsmode fs.FileMode = 0755
)

type daemon struct {
	Mirrors []string `json:"registry-mirrors,omitempty"`
}

type image struct {
	Reg     string
	Mirrors []string
}

func NewImg(reg string, mirror []string) *image {
	for i, m := range mirror {
		mirror[i] = normalMirror(m)
	}
	return &image{
		Reg:     reg,
		Mirrors: mirror,
	}
}

// 追加仓库镜像, 并重启相关进程
func (m *image) SetMirror() error {
	fs, err := procfs.NewFS("/proc")
	if err != nil {
		return err
	}
	procs, err := fs.AllProcs()
	if err != nil {
		return err
	}
	var (
		dockerpids     []int
		containerdpids []int
	)
	for _, p := range procs {
		cmd, err := p.Comm()
		if err != nil {
			continue
		}
		switch cmd {
		case dockerproc:
			dockerpids = append(dockerpids, p.PID)
		case containerdproc:
			containerdpids = append(containerdpids, p.PID)
		default:
			continue
		}
	}
	if len(dockerpids) != 0 {
		log.Printf("发现 dockerd, 仓库镜像 %s->%s\n", "docker.io", m.Mirrors)
		err = m.setDockerd()
		if err != nil {
			return err
		}
		return util.SendSignal(dockerpids[0], syscall.SIGHUP)
	}
	if len(containerdpids) != 0 {
		log.Printf("发现 containerd, 仓库镜像 %s->%s\n", m.Reg, m.Mirrors)
		err = m.setContainerd()
		if err != nil {
			return err
		}
		return util.SendSignal(containerdpids[0], syscall.SIGHUP)
	}
	log.Panicf("未发现容器有关")
	return nil
}

func (m *image) setDockerd() error {
	var (
		daemonjson = "/etc/docker/daemon.json"
		mirror     = "registry-mirrors"
	)

	err := util.CreateIfNotExist(daemonjson, fsmode)

	data, _ := os.ReadFile(daemonjson)

	var jsonData = make(map[string]interface{})
	if len(data) != 0 {
		err = json.Unmarshal(data, &jsonData)
		if err != nil {
			log.Fatalf("Error unmarshalling JSON: %v", err)
		}
	}

	v, ok := jsonData[mirror]
	if !ok {
		jsonData[mirror] = m.Mirrors
	} else {
		mrs, ok := v.([]string)
		if !ok {

		}
		tmp := util.SetAppend(mrs, m.Mirrors)
		if len(mrs) == len(tmp) {
			log.Panicln("仓库已含 '%s', 不处理", m.Mirrors)
			return nil
		}
	}
	out, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}

	return os.WriteFile(daemonjson, out, fsmode)
}

func (m *image) setContainerd() error {
	var (
		configpath = "/etc/containerd/config.toml"
	)
	err := util.CreateIfNotExist(configpath, fsmode)
	if err != nil {
		return err
	}
	var (
		f, _ = os.Open(configpath)
		cfg  = &ccommand.Config{}
	)

	if err = toml.NewDecoder(f).Decode(cfg); err != nil {
		log.Fatalf("decoding TOML '%s' failed: %v", configpath, err)
	}
	f.Close()

	cri, ok := cfg.Plugins["io.containerd.grpc.v1.cri"]
	if !ok {
		log.Fatalf("containerd 配置文件 %s 未包含 cri 配置", configpath)
	}
	var (
		cricfg = cri.(*criconfig.PluginConfig)
	)
	if cricfg.Registry.Mirrors != nil {
		mrs, ok := cricfg.Registry.Mirrors[m.Reg]
		if !ok {
			cricfg.Registry.Mirrors[m.Reg] = criconfig.Mirror{
				Endpoints: m.Mirrors,
			}
		} else {
			tmp := util.SetAppend(mrs.Endpoints, m.Mirrors)
			if len(tmp) == len(cricfg.Registry.Mirrors[m.Reg].Endpoints) {
				log.Panicln("仓库 '%s' 已含 '%s', 不处理", m.Reg, m.Mirrors)
				return nil
			}
			cricfg.Registry.Mirrors[m.Reg] = criconfig.Mirror{Endpoints: tmp}
		}
	}

	f, err = os.OpenFile(configpath, os.O_TRUNC|os.O_RDWR, fsmode)
	if err != nil {
		log.Fatalf("打开 '%s' 失败: %v", configpath, err)
	}
	defer f.Close()
	cfg.WriteTo(f)
	return nil
}

func normalMirror(s string) string {
	if strings.Index(s, "://") < 0 {
		return "https://" + s
	}
	return s
}
