package kenv

import (
	"context"
	"os"
	"strings"
	"syscall"

	"github.com/spelens-gud/Verktyg.git/kits/klog/logger"
)

func ForkAndInheritEnv(env []string) (pid int, err error) {
	var (
		ctx = context.Background()
		lg  = logger.FromContext(ctx)
	)
	lg.Infof("listening to socket fd from ppid [ %d ] ", syscall.Getppid())
	f := os.NewFile(3, "")
	defer func() {
		_ = f.Close()
	}()
	var (
		osEnv = os.Environ()
		envs  = make([]string, 0, len(osEnv))
	)
	for _, osValue := range osEnv {
		for _, v := range env {
			split := strings.Split(v, "=")
			if !strings.HasPrefix(osValue, split[0]) {
				envs = append(envs, osValue)
			}
		}
	}
	envs = append(envs, env...)
	execSpec := &syscall.ProcAttr{
		Env:   envs,
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd(), f.Fd()},
	}
	pid, _, err = syscall.StartProcess(os.Args[0], os.Args, execSpec)
	return
}
