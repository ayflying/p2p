//go:build windows

package os

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/getlantern/systray"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
)

// 引入 Windows API 函数
var (
	user32        = syscall.NewLazyDLL("user32.dll")
	kernel32      = syscall.NewLazyDLL("kernel32.dll")
	showWindow    = user32.NewProc("ShowWindow")
	getConsoleWnd = kernel32.NewProc("GetConsoleWindow")
)

func (s *sOS) start() {

	// 系统托盘初始化（设置图标、右键菜单）
	go systray.Run(s.onSystrayReady, s.onSystrayExit)
}

// 系统托盘初始化（设置图标、右键菜单）
func (s *sOS) onSystrayReady() {
	//s.hideConsole()

	iconByte := gfile.GetBytes(s.systray.Icon)
	systray.SetIcon(iconByte)
	systray.SetTitle(s.systray.Title)
	systray.SetTooltip(s.systray.Tooltip)

	mQuit := systray.AddMenuItem("退出", "退出应用")
	mShow := systray.AddMenuItemCheckbox("显示窗口", "显示窗口", false)
	// Sets the icon of a menu item. Only available on Mac and Windows.
	//mQuit.SetIcon(iconByte)
	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				systray.Quit()
			case <-mShow.ClickedCh:
				// 显示窗口
				s.showConsole()
			}

		}
	}()

}

func (s *sOS) onSystrayExit() {
	// clean up here
	g.Log().Debugf(gctx.New(), "系统托盘退出")
	defer os.Exit(0)

}

func (s *sOS) update(version, server string) {
	updaterPath := gcmd.GetArg(0).String()
	// 启动更新器，传递主程序路径和当前版本作为参数
	cmd := exec.Command(updaterPath, version, server)
	// 将更新器与主程序的控制台分离
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: false} // Windows特定设置

	if err := cmd.Start(); err != nil {
		return
	}
}

// 隐藏控制台窗口
func (s *sOS) hideConsole() {
	// 获取当前控制台窗口句柄
	hWnd, _, _ := getConsoleWnd.Call()
	if hWnd == 0 {
		return // 无控制台窗口（如编译为GUI子系统时）
	}
	// SW_HIDE = 0：隐藏窗口
	showWindow.Call(hWnd, 0)
}

// 显示控制台窗口
func (s *sOS) showConsole() {
	// 获取当前控制台窗口句柄
	hWnd, _, _ := getConsoleWnd.Call()
	if hWnd == 0 {
		return
	}
	// SW_SHOW = 5：显示窗口
	showWindow.Call(hWnd, 5)
}
