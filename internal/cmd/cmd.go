package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/ayflying/p2p/internal/service"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
)

func init() {
	err := Main.AddCommand(&Main, &Debug, &P2p, &DHT)
	if err != nil {
		g.Log().Error(gctx.GetInitCtx(), err)
		return
	}
}

var (
	s = g.Server()

	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			g.Log().Debug(ctx, "开始执行main")

			parser, err = gcmd.Parse(g.MapStrBool{
				"p,port": true,
			})
			port := parser.GetOpt("port", "23333").Int()

			h, _ := service.P2P().CreateLibp2pHost(gctx.New(), port)
			err = service.P2P().DHTStart(h, nil)
			if err != nil {
				g.Log().Error(ctx, err)
			}

			time.Sleep(5 * time.Second)
			publicIp, _ := service.P2P().GetIPv4PublicIP()
			validKey := fmt.Sprintf("%v/ip", h.ID())
			dataValue := fmt.Sprintf("来自节点 %s 的数据:%v", h.ID().ShortString(), publicIp)
			if err = service.P2P().StoreToDHT(gctx.New(), validKey, dataValue); err != nil {
				g.Log().Debugf(ctx, "❌ 存储失败: %v\n", err)
			} else {
				g.Log().Debugf(ctx, "✅ 存储成功\nKey: %s\nValue: %s\n", validKey, dataValue)
			}

			//parser, err = gcmd.Parse(g.MapStrBool{
			//	"w,ws":      true,
			//	"g,gateway": true,
			//	"p,port":    true,
			//})
			//addr := g.Cfg().MustGet(ctx, "ws.address").String()
			//ws := parser.GetOpt("ws", addr).String()
			////port := parser.GetOpt("port", 0).Int()
			//
			//s.Group("/", func(group *ghttp.RouterGroup) {
			//	group.Middleware(ghttp.MiddlewareHandlerResponse)
			//	group.Bind(
			//		p2p.NewV1(),
			//	)
			//})
			//
			////启动p2p服务端网关
			//s.Group("/ws", func(group *ghttp.RouterGroup) {
			//	group.Middleware(ghttp.MiddlewareHandlerResponse)
			//	err = service.P2P().GatewayStart(ctx, group)
			//	if err != nil {
			//		g.Log().Error(ctx, err)
			//	}
			//})
			//
			////s.SetPort(port)
			//
			//// 延迟启动
			//gtimer.SetTimeout(ctx, time.Second*5, func(ctx context.Context) {
			//	g.Log().Debug(ctx, "开始执行客户端")
			//	// 启动p2p客户端
			//	err = service.P2P().Start(ws)
			//
			//	g.Log().Debugf(ctx, "当前监听端口:%v", s.GetListenedPort())
			//	//addrs, _ := net.InterfaceAddrs()
			//	//for _, addr := range addrs {
			//	//	if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			//	//		g.Log().Infof(ctx, "访问地址:http://%v:%d", ipnet.IP.String(), s.GetListenedPort())
			//	//	}
			//	//}
			//
			//})

			s.Run()
			return nil
		},
	}
)
