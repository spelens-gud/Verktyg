package kdoc

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var moduleRegexp = regexp.MustCompile("module (.+)")

var modPath string
var modBase string

func getModPath() string {
	if len(modPath) > 0 {
		return modPath
	}
	cmd := exec.Command("go", "env", "GOMOD")
	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout
	_ = cmd.Run()
	out := stdout.String()
	modPath = strings.Trim(out, "\n")
	return modPath
}

func getModBase() string {
	if len(modBase) > 0 {
		return modBase
	}
	m, err := os.ReadFile(getModPath())
	if err != nil {
		return ""
	}
	f := moduleRegexp.FindStringSubmatch(strings.Split(string(m), "\n")[0])
	if len(f) == 2 {
		base := f[1]
		modBase = base
		return base
	}
	return ""
}

func getImportCodePath(importPath string) (serviceCodePath string, ok bool) {
	modBase := getModBase()
	if len(modBase) > 0 && strings.Contains(importPath, modBase) {
		// 根据go.mod定位
		modPath := getModPath()
		modDir := filepath.Dir(filepath.ToSlash(modPath))
		serviceCodePath = strings.ReplaceAll(importPath, modBase, modDir)
		ok = true
		return
	}
	return
}
