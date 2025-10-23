package system

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ayflying/p2p/api/system/v1"
	"github.com/gogf/gf/v2/crypto/gsha1"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
)

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {

	getRunFile := gcmd.GetArg(0).String()

	fileSha, err := gsha1.EncryptFile(getRunFile)
	g.Dump(fileSha)
	g.Dump(getRunFile)

	go func() {
		log.Println("5秒后开始重启...")
		time.Sleep(5 * time.Second)

		if err = restartSelf(); err != nil {
			log.Fatalf("重启失败：%v", err)
		}
	}()

	return
}

// restartSelf 实现 Windows 平台下的程序自重启
func restartSelf() error {
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
