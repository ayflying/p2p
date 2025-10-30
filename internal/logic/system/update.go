package system

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ayflying/p2p/internal/service"
	"github.com/gogf/gf/v2/encoding/gcompress"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
)

// 本地版本号（建议从编译参数注入，如 -ldflags "-X main.version=v0.1.3"）
var localVersion = "v0.0.0"

// 对应 GitHub API 响应的核心字段（按需精简）
type GitHubRelease struct {
	Url             string    `json:"url"`
	AssetsUrl       string    `json:"assets_url"`
	UploadUrl       string    `json:"upload_url"`
	HtmlUrl         string    `json:"html_url"`
	Id              int       `json:"id"`
	TagName         string    `json:"tag_name"`
	Assets          []*Assets `json:"assets"`
	NodeId          string    `json:"node_id"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Draft           bool      `json:"draft"`
	Immutable       bool      `json:"immutable"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	PublishedAt     time.Time `json:"published_at"`
	TarballUrl      string    `json:"tarball_url"`
	ZipballUrl      string    `json:"zipball_url"`
	Body            string    `json:"body"`
}

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

// RenameRunningFile 重命名正在运行的程序文件（如 message.exe → message.exe~）
func (s *sSystem) RenameRunningFile(exePath string) (string, error) {
	// 目标备份文件名（message.exe → message.exe~）
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

// 简化版版本对比（仅适用于 vX.Y.Z 格式）
func (s *sSystem) isNewVersion(local, latest string) bool {
	// 移除前缀 "v"，按 "." 分割成数字切片
	localParts := strings.Split(strings.TrimPrefix(local, "v"), ".")
	latestParts := strings.Split(strings.TrimPrefix(latest, "v"), ".")

	// 逐段对比版本号（如 0.1.3 vs 0.1.4 → 后者更新）
	for i := 0; i < len(localParts) && i < len(latestParts); i++ {
		if localParts[i] < latestParts[i] {
			return true
		} else if localParts[i] > latestParts[i] {
			return false
		}
	}
	// 若前缀相同，长度更长的版本更新（如 0.1 vs 0.1.1）
	return len(localParts) < len(latestParts)
}

func (s *sSystem) getLatestVersion() (string, []*Assets, error) {
	apiURL := "https://api.github.com/repos/ayflying/p2p/releases/latest"
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", nil, fmt.Errorf("请求失败：%v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("API 响应错误：%d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", nil, fmt.Errorf("解析响应失败：%v", err)
	}

	return release.TagName, release.Assets, nil
}

func (s *sSystem) CheckUpdate() {
	ctx := gctx.New()
	latestVersion, assets, err := s.getLatestVersion()
	if err != nil {
		fmt.Printf("检查更新失败：%v\n", err)
		return
	}

	localVersion = gfile.GetContents("download/version.txt")

	if s.isNewVersion(localVersion, latestVersion) {
		g.Log().Printf(ctx, "发现新版本：%s（当前版本：%s）", latestVersion, localVersion)
		//拼接操作系统和架构（格式：OS_ARCH）
		platform := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
		//name := fmt.Sprintf("p2p_%s_%s.tar.gz", latestVersion, platform)
		fmt.Println("下载链接：")
		for _, asset := range assets {
			if strings.Contains(fmt.Sprintf("_%s.", asset.Name), platform) {
				fmt.Printf("- %s\n", asset.BrowserDownloadUrl)

				fileDownload, err2 := g.Client().Get(ctx, asset.BrowserDownloadUrl)
				if err2 != nil {
					return
				}
				//filename := gfile.Name()
				err = gfile.PutBytes(path.Join("download", asset.Name), fileDownload.ReadAll())

				// 保存最新版本号到文件
				gfile.PutContents("download/version.txt", latestVersion)
				break
			}
		}
	} else {
		fmt.Printf("当前已是最新版本：%s\n", localVersion)
	}
}
