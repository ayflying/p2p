package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

var (
	updateServer = "https://your-server.com/update"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("用法: updater <主程序路径> <当前版本> <更新服务器URL>")
		os.Exit(1)
	}

	mainExePath := os.Args[1]
	currentVersion := os.Args[2]
	updateServer = os.Args[3]

	fmt.Println("开始更新程序...")

	// 1. 等待主程序完全退出
	if err := waitForMainExit(mainExePath); err != nil {
		fmt.Printf("等待主程序退出失败: %v\n", err)
		pauseAndExit()
	}

	// 2. 获取最新版本信息
	updateInfo, err := getLatestVersionInfo(currentVersion)
	if err != nil {
		fmt.Printf("获取更新信息失败: %v\n", err)
		pauseAndExit()
	}

	// 3. 下载新版本
	newExePath, err := downloadNewVersion(updateInfo.DownloadURL)
	if err != nil {
		fmt.Printf("下载更新失败: %v\n", err)
		pauseAndExit()
	}

	// 4. 替换主程序
	if err := replaceMainExe(mainExePath, newExePath); err != nil {
		fmt.Printf("替换程序失败: %v\n", err)
		pauseAndExit()
	}

	// 5. 重启主程序
	if err := restartMainProgram(mainExePath); err != nil {
		fmt.Printf("重启程序失败: %v\n", err)
		pauseAndExit()
	}

	fmt.Println("更新完成")
}

// 等待主程序退出
func waitForMainExit(mainPath string) error {
	// 获取主程序文件名
	mainName := filepath.Base(mainPath)

	for {
		// 检查是否还有同名进程在运行
		if !isProcessRunning(mainName) {
			return nil
		}
		time.Sleep(300 * time.Millisecond)
	}
}

// 检查进程是否在运行
func isProcessRunning(name string) bool {
	// Windows下通过tasklist命令检查进程
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s", name))
	_, err := cmd.Output()
	if err != nil {
		return false
	}

	return true
}

// 获取最新版本信息
func getLatestVersionInfo(currentVersion string) (UpdateInfo, error) {
	var info UpdateInfo
	resp, err := http.Get(fmt.Sprintf("%s/latest?current=%s", updateServer, currentVersion))
	if err != nil {
		return info, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return info, err
	}

	return info, nil
}

// 下载新版本
func downloadNewVersion(url string) (string, error) {
	tempFile := filepath.Join(os.TempDir(), "main_new.exe")
	out, err := os.Create(tempFile)
	if err != nil {
		return "", err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 下载并计算哈希值
	hash := sha256.New()
	writer := io.MultiWriter(out, hash)

	if _, err := io.Copy(writer, resp.Body); err != nil {
		os.Remove(tempFile)
		return "", err
	}

	// 这里可以添加哈希校验逻辑

	return tempFile, nil
}

// 替换主程序文件
func replaceMainExe(oldPath, newPath string) error {
	// 备份旧版本
	backupPath := oldPath + ".bak"
	if err := os.Remove(backupPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	if err := os.Rename(oldPath, backupPath); err != nil {
		return err
	}

	// 移动新版本到目标路径
	if err := os.Rename(newPath, oldPath); err != nil {
		// 替换失败，恢复备份
		os.Rename(backupPath, oldPath)
		return err
	}

	// 替换成功，删除备份
	os.Remove(backupPath)
	return nil
}

// 重启主程序
func restartMainProgram(mainPath string) error {
	cmd := exec.Command(mainPath)
	// 在后台启动主程序
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: false}
	return cmd.Start()
}

// 暂停并退出（给用户看错误信息）
func pauseAndExit() {
	fmt.Println("按任意键退出...")
	var input string
	fmt.Scanln(&input)
	os.Exit(1)
}

// 更新信息结构体
type UpdateInfo struct {
	Version     string `json:"version"`
	DownloadURL string `json:"download_url"`
	SHA256      string `json:"sha256"`
}
