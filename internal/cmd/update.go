package cmd

import (
	"context"
	"fmt"
	"path"
	"runtime"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gcmd"
)

var (
	Update = gcmd.Command{
		Name:  "update",
		Usage: "update",
		Brief: "更新版本",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			g.Log().Info(ctx, "准备上传更新文件")
			//加载编辑配置文件
			g.Cfg("hack").GetAdapter().(*gcfg.AdapterFile).SetFileName("hack/config.yaml")
			//获取文件名
			getFileName, err := g.Cfg("hack").Get(ctx, "gfcli.build.name")
			filename := getFileName.String()

			getPath, err := g.Cfg("hack").Get(ctx, "gfcli.build.path")
			pathMain := getPath.String()

			//获取版本号
			getVersion, err := g.Cfg("hack").Get(ctx, "gfcli.build.version")
			version := getVersion.String()

			// 拼接操作系统和架构（格式：OS_ARCH）
			platform := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)

			var filePath = path.Join(pathMain, version, platform, filename)

			g.Log().Debugf(ctx, "当前获取到的地址为：%v", filePath)
			return
		}}
)
