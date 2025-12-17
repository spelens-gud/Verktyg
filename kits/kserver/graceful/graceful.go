package graceful

import (
	"errors"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/spelens-gud/Verktyg/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg/kits/kenv/envflag"
)

const envGracefulReloadFlag = "GRACEFUL_RELOAD_FLAG"

var (
	// 设置重启信号
	enableGracefulReload = envflag.BoolFrom(iconfig.EnvKeyServerGracefulReload)
	reloadFlag           = envflag.BoolFrom(envGracefulReloadFlag)
)

type fileListener interface {
	File() (f *os.File, err error)
}

func IsGracefulChild() bool {
	return reloadFlag()
}

func EnvGracefulReload() bool {
	return enableGracefulReload()
}

func ForkWithListener(listener net.Listener) (fork int, err error) {
	tl, ok := listener.(fileListener)
	if !ok {
		err = errors.New("listener is not valid tcp listener")
		return
	}

	f, err := tl.File()
	if err != nil {
		return
	}

	var (
		osEnv = os.Environ()
		envs  = make([]string, 0, len(osEnv))
		flag  = envGracefulReloadFlag + "=1"
	)

	for _, value := range osEnv {
		if !strings.HasPrefix(value, envGracefulReloadFlag+"=") {
			envs = append(envs, value)
		}
	}

	envs = append(envs, flag)

	_ = os.Chmod(os.Args[0], 0775)

	fork, _, err = syscall.StartProcess(os.Args[0], os.Args, &syscall.ProcAttr{
		Env:   envs,
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd(), f.Fd()},
	})
	return
}
