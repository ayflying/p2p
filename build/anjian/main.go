package main

import (
	"syscall"
	"time"
	"unsafe"
)

// 定义Windows API所需的结构体和常量（对应SendInput函数参数）
const (
	INPUT_MOUSE          = 0x0000 // 输入类型：鼠标
	MOUSEEVENTF_LEFTDOWN = 0x0002 // 左键按下
	MOUSEEVENTF_LEFTUP   = 0x0004 // 左键释放
)

// INPUT结构体：SendInput的输入参数
type INPUT struct {
	Type uint32
	Mi   MOUSEINPUT
}

// MOUSEINPUT结构体：鼠标输入详情
type MOUSEINPUT struct {
	Dx          int32
	Dy          int32
	MouseData   uint32
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

func main() {
	// 加载user32.dll并获取SendInput函数地址
	user32, err := syscall.LoadLibrary("user32.dll")
	if err != nil {
		panic(err)
	}
	defer syscall.FreeLibrary(user32)

	sendInputProc, err := syscall.GetProcAddress(user32, "SendInput")
	if err != nil {
		panic(err)
	}

	// 定义一次完整点击的输入（按下+释放）
	inputs := []INPUT{
		{
			Type: INPUT_MOUSE,
			Mi: MOUSEINPUT{
				DwFlags: MOUSEEVENTF_LEFTDOWN, // 左键按下
			},
		},
		{
			Type: INPUT_MOUSE,
			Mi: MOUSEINPUT{
				DwFlags: MOUSEEVENTF_LEFTUP, // 左键释放
			},
		},
	}

	// 循环执行1000次点击，间隔10毫秒
	for i := 0; i < 1000; i++ {
		// 调用SendInput发送鼠标事件（参数：输入数量、输入数组指针、单个输入大小）
		var args []uintptr
		args = append(args, uintptr(len(inputs)))
		args = append(args, uintptr(unsafe.Pointer(&inputs[0])))
		args = append(args, uintptr(unsafe.Sizeof(INPUT{})))

		ret, _, _ := syscall.SyscallN(sendInputProc, args...)
		if ret == 0 {
			panic("发送鼠标事件失败")
		}

		// 间隔10毫秒
		time.Sleep(10 * time.Millisecond)
	}

	println("已完成1000次鼠标点击")
}
