package system

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ayflying/p2p/internal/service"
	"github.com/gogf/gf/v2/encoding/gcompress"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
)

func (s *sSystem) Update(ctx context.Context) (err error) {
	//拼接操作系统和架构（格式：OS_ARCH）
	platform := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)

	runFile := gcmd.GetArg(0).String()
	oldFile, err := service.System().RenameRunningFile(runFile)
	g.Log().Debugf(ctx, "执行文件改名为%v", oldFile)
	gz := path.Join("download", platform+".gz")
	err = gcompress.UnGzipFile(gz, runFile)

	go func() {
		log.Println("5秒后开始重启...")
		time.Sleep(5 * time.Second)

		if err = service.System().RestartSelf(); err != nil {
			log.Fatalf("重启失败：%v", err)
		}
	}()
	return
}

// RestartSelf 实现 Windows 平台下的程序自重启
func (s *sSystem) RestartSelf() error {
	// 1. 获取当前程序的绝对路径
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	// 处理路径中的符号链接（确保路径正确）
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return err
	}

	// 2. 获取命令行参数（os.Args[0] 是程序名，实际参数从 os.Args[1:] 开始）
	args := os.Args[1:]

	// 3. 构建新进程命令（路径为当前程序，参数为原参数）
	cmd := exec.Command(exePath, args...)
	// 设置新进程的工作目录与当前进程一致
	cmd.Dir, err = os.Getwd()
	if err != nil {
		return err
	}

	// 新进程的输出继承当前进程的标准输出（可选，根据需求调整）
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// 4. 启动新进程（非阻塞，Start() 后立即返回）
	if err := cmd.Start(); err != nil {
		return err
	}

	// 5. 新进程启动成功后，退出当前进程
	os.Exit(0)
	return nil // 理论上不会执行到这里
}

// RenameRunningFile 重命名正在运行的程序文件（如 p2p.exe → p2p.exe~）
func (s *sSystem) RenameRunningFile(exePath string) (string, error) {
	// 目标备份文件名（p2p.exe → p2p.exe~）
	backupPath := exePath + "~"

	// 先删除已存在的备份文件（若有）
	if _, err := os.Stat(backupPath); err == nil {
		if err := os.Remove(backupPath); err != nil {
			return "", fmt.Errorf("删除旧备份文件失败: %v", err)
		}
	}

	// 重命名正在运行的 exe 文件
	// 关键：Windows 允许对锁定的文件执行重命名操作
	if err := os.Rename(exePath, backupPath); err != nil {
		return "", fmt.Errorf("重命名运行中文件失败: %v", err)
	}
	return backupPath, nil
}
