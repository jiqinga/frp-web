//go:build !windows

/*
 * @Author              : 寂情啊
 * @Date                : 2026-01-06 17:25:02
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2026-01-06 17:25:15
 * @FilePath            : frp-web-testbackendinternalserviceprocess_manager_unix.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */

package service

import (
	"os/exec"
	"syscall"
)

func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
