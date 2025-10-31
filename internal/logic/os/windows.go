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
	"github.com/gogf/gf/v2/os/gres"
)

// 引入 Windows API 函数
var (
	user32                = syscall.NewLazyDLL("user32.dll")
	kernel32              = syscall.NewLazyDLL("kernel32.dll")
	showWindow            = user32.NewProc("ShowWindow")
	getConsoleWnd         = kernel32.NewProc("GetConsoleWindow")
	freeConsole           = kernel32.NewProc("FreeConsole")
	attachConsole         = kernel32.NewProc("AttachConsole")
	allocConsole          = kernel32.NewProc("AllocConsole")
	setConsoleCtrlHandler = kernel32.NewProc("SetConsoleCtrlHandler")
	consoleCtrlHandler    uintptr
)

const (
	ctrlCloseEvent = 2
)

func (s *sOS) start() {
	// 注册控制台关闭事件处理：点击叉叉仅隐藏控制台而不退出程序
	s.setupConsoleCloseHandler()

	// 系统托盘初始化（设置图标、右键菜单）
	go systray.Run(s.onSystrayReady, s.onSystrayExit)
}

// 系统托盘初始化（设置图标、右键菜单）
func (s *sOS) onSystrayReady() {
	// s.hideConsole()
	var iconByte []byte
	if !gfile.Exists(s.systray.Icon) {
		iconByte = gres.GetContent(s.systray.Icon)
		gfile.PutBytes(s.systray.Icon, iconByte)
	}
	iconByte = gfile.GetBytes(s.systray.Icon)

	//if gres.Contains(s.systray.Icon) {
	//	iconByte = gres.GetContent(s.systray.Icon)
	//	gfile.PutBytes(s.systray.Icon, iconByte)
	//} else {
	//	iconByte = gfile.GetBytes(s.systray.Icon)
	//}

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
	// 仅隐藏控制台窗口（保留现有缓冲区以便后续显示时保留历史日志）
	hWnd, _, _ := getConsoleWnd.Call()
	if hWnd != 0 {
		// SW_HIDE = 0：隐藏窗口
		showWindow.Call(hWnd, 0)
	}
}

// 显示控制台窗口
func (s *sOS) showConsole() {
	// 获取当前控制台窗口句柄
	hWnd, _, _ := getConsoleWnd.Call()
	if hWnd == 0 {
		// 如果当前进程没有控制台，尝试附加到父进程控制台
		// ATTACH_PARENT_PROCESS = (DWORD)-1
		ret, _, _ := attachConsole.Call(uintptr(^uint32(0)))
		if ret == 0 {
			// 附加失败则分配一个新的控制台窗口
			allocConsole.Call()
		}
		// 重新获取控制台窗口句柄
		hWnd, _, _ = getConsoleWnd.Call()
	}
	if hWnd != 0 {
		// SW_SHOW = 5：显示窗口
		showWindow.Call(hWnd, 5)
	}
}

// 注册控制台关闭事件处理器，将关闭事件转换为隐藏行为
func (s *sOS) setupConsoleCloseHandler() {
	if consoleCtrlHandler != 0 {
		return
	}
	consoleCtrlHandler = syscall.NewCallback(func(ctrlType uint32) uintptr {
		if ctrlType == ctrlCloseEvent {
			// 用户点击控制台窗口的关闭按钮(X)：仅隐藏，不退出
			hWnd, _, _ := getConsoleWnd.Call()
			if hWnd != 0 {
				// SW_HIDE = 0
				showWindow.Call(hWnd, 0)
			}
			// 返回 TRUE 表示事件已处理，阻止默认终止行为
			return 1
		}
		// 其他事件交由系统默认处理
		return 0
	})
	setConsoleCtrlHandler.Call(consoleCtrlHandler, uintptr(1))
}
