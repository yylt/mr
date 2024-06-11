package util

import (
	"io/fs"
	"os"
	"path"
	"runtime"
	"syscall"
)

type Os string

const (
	Linux  Os = "linux"
	Window Os = "windows"
	Mac    Os = "darwin"
)

func IsSupportOs(oss ...string) bool {
	for _, v := range oss {
		if v == runtime.GOOS {
			return true
		}
	}
	return false
}

func CreateIfNotExist(fpath string, mod fs.FileMode) error {
	var (
		dir = path.Dir(fpath)
	)
	_, err := os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		os.Mkdir(dir, mod)
	}

	_, err = os.Stat(fpath)
	if err != nil && os.IsNotExist(err) {
		fs, err := os.OpenFile(dir, os.O_CREATE, mod)
		if err != nil {
			return err
		}
		return fs.Close()
	}
	return nil
}

func SendSignal(pid int, sig syscall.Signal) error {
	return syscall.Kill(pid, sig)
}
