// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

type (
	IOS interface {
		Load(title string, tooltip string, ico string)
		Update(version string, server string)
	}
)

var (
	localOS IOS
)

func OS() IOS {
	if localOS == nil {
		panic("implement not found for interface IOS, forgot register?")
	}
	return localOS
}

func RegisterOS(i IOS) {
	localOS = i
}
