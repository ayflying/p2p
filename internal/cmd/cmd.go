package cmd

import (
	"context"
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
				"p,port": true,
			})
			//port := parser.GetOpt("port", "23333").Int()

			parser, err = gcmd.Parse(g.MapStrBool{
				"w,ws":      true,
				"g,gateway": true,
				"p,port":    true,
			})
			addr := g.Cfg().MustGet(ctx, "ws.address").String()
			ws := parser.GetOpt("ws", addr).String()
			//port := parser.GetOpt("port", 0).Int()

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

			//s.SetPort(port)

			// 延迟启动
			gtimer.SetTimeout(ctx, time.Second*5, func(ctx context.Context) {
				g.Log().Debug(ctx, "开始执行客户端")
				// 启动p2p客户端
				err = service.P2P().Start(ws)

				g.Log().Debugf(ctx, "当前监听端口:%v", s.GetListenedPort())
				//addrs, _ := net.InterfaceAddrs()
				//for _, addr := range addrs {
				//	if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				//		g.Log().Infof(ctx, "访问地址:http://%v:%d", ipnet.IP.String(), s.GetListenedPort())
				//	}
				//}

			})

			// 启动系统托盘
			service.OS().Load(consts.Name, consts.Name+"服务端", "manifest/images/favicon.ico")

			s.Run()
			return nil
		},
	}
)
