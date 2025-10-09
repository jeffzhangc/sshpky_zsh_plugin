package sshrunner

import (
	"errors"
	"runtime"

	"github.com/riywo/loginshell"
)

func getShell() (string, error) {
	switch runtime.GOOS {
	case "plan9":
		return loginshell.Plan9Shell()
	case "linux":
		return "/bin/bash", nil
	case "openbsd":
		return loginshell.NixShell()
	case "freebsd":
		return loginshell.NixShell()
	case "android":
		return loginshell.AndroidShell()
	case "darwin":
		return "/bin/bash", nil
	case "windows":
		return loginshell.WindowsShell()
	}

	return "", errors.New("Undefined GOOS: " + runtime.GOOS)
}
