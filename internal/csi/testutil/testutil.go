// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package testutil

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

func MakeFakeExecCommand(exitStatus int, stdout string) func(string, ...string) *exec.Cmd {
	return func(command string, args ...string) *exec.Cmd {
		cmd := exec.Command("sh", "-c", `echo -n "$STD_OUT"; exit $EXIT_STATUS`)
		cmd.Env = []string{
			"STD_OUT=" + stdout,
			"EXIT_STATUS=" + strconv.Itoa(exitStatus),
		}
		return cmd
	}
}

func MakeFakeExecCommandContext(exitStatus int, stdout string) func(context.Context, string, ...string) *exec.Cmd {
	return func(ctx context.Context, command string, args ...string) *exec.Cmd {
		return MakeFakeExecCommand(exitStatus, stdout)(command, args...)
	}
}

func MakeFakeExecWithTimeout(withTimeout bool, output []byte, err error) func(string, []string, time.Duration, *syscall.SysProcAttr) ([]byte, error) {
	return func(command string, args []string, timeout time.Duration, sysProcAttr *syscall.SysProcAttr) ([]byte, error) {
		if withTimeout {
			return nil, context.DeadlineExceeded
		}
		return output, err
	}
}

// GetWorkDirPath returns the path to the current working directory.
func GetWorkDirPath(dir string) (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%c%s", path, os.PathSeparator, dir), nil
}

func MakeDir(pathname string) error {
	err := os.MkdirAll(pathname, os.FileMode(0755))
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
	}
	return nil
}
