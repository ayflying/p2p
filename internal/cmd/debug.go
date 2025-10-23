package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/dop251/goja"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"
)

type DebugType struct {
	Action string `json:"action" dc:"操作渠道"`
	Number int64  `json:"number" dc:"数字"`
	String string `json:"string" dc:"字符串"`
	Json   string `json:"json" dc:"符合参数"`
}

var (
	Debug = gcmd.Command{
		Name:  "debug",
		Usage: "debug",
		Brief: "调试接口",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			g.Log().Debug(ctx, "开始执行debug v1.0.5")

			g.Log().SetConfigWithMap(g.Map{
				"level":  "all",
				"stdout": true,
			})
			var req = &DebugType{
				Action: parser.GetOpt("a").String(),
				Number: parser.GetOpt("n").Int64(),
				String: parser.GetOpt("s").String(),
				Json:   parser.GetOpt("j").String(),
			}

			g.Log().Debug(ctx, "开始调试了")
			g.Log().Debugf(ctx, "开始执行:action:%v,number=%v,string=%v,json=%v", req.Action, req.Number, req.String, req.Json)
			var msg any
			switch req.Action {
			case "js":
				vm := goja.New()

				if req.String == "" {
					req.String = "console.log('hello world');"
				}

				res, err := vm.RunString(req.String)
				if err != nil {
					break
				}
				msg = res.Export()
				g.Dump(res.ToNumber())
			case "p2p":
			// host, err := service.P2P().Start(ctx)
			// if err != nil {
			// 	break
			// }
			// g.Dump(host.ID().String(), host.Addrs())
			case "update":
				url := "http://ay.cname.com:5244/d/guest/novapps/%E5%88%86%E5%B8%83%E5%BC%8F/p2p/p2p.exe?sign=8anTHvfJKJLCfZTI4IuopNK38x9rEoDiNevr5aZZPgM=:0"
				g.Log().Debugf(ctx, "当前开始更新了,url=%v", url)
				//service.OS().Update("v1.0.0", "http://127.0.0.1:8080")

				resp, err := g.Client().Get(ctx, url)
				if err != nil {
					g.Log().Error(ctx, err)
				}
				//filename := g.Cfg("hack").MustGet(ctx, "gfcli.build.name").String()
				filename := gcmd.GetArg(0).String()

				_, err = renameRunningFile(filename)
				if err != nil {
					g.Log().Error(ctx, err)
				}

				//switch runtime.GOOS {
				//case "windows":
				//	fmt.Println("当前系统：Windows")
				//	filename = filename + ".exe"
				//	if gfile.Exists(filename) {
				//		filename += "~"
				//	}
				//default:
				//	fmt.Println("当前系统：" + runtime.GOOS)
				//}
				//if gfile.Exists(filename) {
				//	filename += "~"
				//}
				err = gfile.PutBytes(filename, resp.ReadAll())
				if err != nil {
					g.Log().Error(ctx, err)
				}
				msg = "下载完成了"
			}
			g.Log().Debug(ctx, msg)
			return
		},
	}
)

// 重命名正在运行的程序文件（如 p2p.exe → p2p.exe~）
func renameRunningFile(exePath string) (string, error) {
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
