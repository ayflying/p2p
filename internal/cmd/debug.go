package cmd

import (
	"context"

	"github.com/ayflying/p2p/internal/service"
	"github.com/dop251/goja"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
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
			g.Log().Debug(ctx, "开始执行debug")

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
				service.OS().Update("v1.0.0", "http://127.0.0.1:8080")
			}
			g.Log().Debug(ctx, msg)
			return
		},
	}
)
