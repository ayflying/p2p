// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
)

type (
	ISystem interface {
		Init()
		Update(ctx context.Context) (err error)
		// RestartSelf 实现 Windows 平台下的程序自重启
		RestartSelf() error
		// RenameRunningFile 重命名正在运行的程序文件（如 message.exe → message.exe~）
		RenameRunningFile(exePath string) (string, error)
	}
)

var (
	localSystem ISystem
)

func System() ISystem {
	if localSystem == nil {
		panic("implement not found for interface ISystem, forgot register?")
	}
	return localSystem
}

func RegisterSystem(i ISystem) {
	localSystem = i
}
