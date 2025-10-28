package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/ayflying/p2p/internal/consts"
	"github.com/ayflying/p2p/internal/controller/p2p"
	"github.com/ayflying/p2p/internal/controller/system"
	"github.com/ayflying/p2p/internal/service"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gtimer"
	"github.com/gogf/gf/v2/util/grand"
)

func init() {
	err := Main.AddCommand(&Main, &Debug, &Update)
	if err != nil {
		g.Log().Error(gctx.GetInitCtx(), err)
		return
	}
}

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			g.Log().Debug(ctx, "开始执行main")
			Version, err := g.Cfg("hack").Get(ctx, "gfcli.build.version")
			g.Log().Debugf(ctx, "当前启动的版本为：%v", Version)

			s := g.Server(consts.Name)

			parser, err = gcmd.Parse(g.MapStrBool{
				"w,ws":      true,
				"g,gateway": true,
				"p,port":    true,
				"t,type":    true,
			})
			//addr := g.Cfg().MustGet(ctx, "ws.address").String()
			ws := parser.GetOpt("ws").String()
			if ws == "" {
				listVar := g.Cfg().MustGet(ctx, "p2p.list")
				var p2pItem []struct {
					Host string `json:"host"`
					Port int    `json:"port"`
					SSL  bool   `json:"ssl"`
					Ws   string `json:"ws"`
				}
				listVar.Scan(&p2pItem)
				key := grand.Intn(len(p2pItem) - 1)
				wsData := p2pItem[key]
				ws = fmt.Sprintf("ws://%s:%d/ws", wsData.Host, wsData.Port)
			}

			port := parser.GetOpt("port", 0).Int()
			s.SetPort(port)
			//if port > 0 {
			//	s.SetPort(port)
			//}

			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(
					p2p.NewV1(),
					system.NewV1(),
				)
			})

			//启动p2p服务端网关
			s.Group("/ws", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				err = service.P2P().GatewayStart(ctx, group)
				if err != nil {
					g.Log().Error(ctx, err)
				}
			})

			startType := parser.GetOpt("type").String()
			if startType != "server" {
				// 延迟启动
				gtimer.SetTimeout(ctx, time.Second*5, func(ctx context.Context) {
					g.Log().Debug(ctx, "开始执行客户端")
					// 启动p2p客户端
					err = service.P2P().Start(ws)

					g.Log().Debugf(ctx, "当前监听端口:%v", s.GetListenedPort())
				})
				//s.SetPort(0)
			}

			// 启动系统托盘
			service.OS().Load(consts.Name, consts.Name+"服务端", "manifest/images/favicon.ico")

			s.Run()
			return nil
		},
	}
)
