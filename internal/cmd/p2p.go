package cmd

import (
	"context"

	"github.com/ayflying/p2p/internal/service"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
)

// p2pHelpDescription 定义P2P命令的详细帮助信息
const p2pHelpDescription = `
P2P连接工具使用帮助:

模式1: 网关服务器
  功能: 拥有外网IP，接收客户端连接，协助P2P打洞
  命令: p2p -a gateway

模式2: 客户端
  功能: 连接到网关，通过打洞实现与其他客户端的长连接通讯
  命令: p2p -a client --gateway 网关ID

高级功能:
  客户端间连接: p2p --mode client --gateway 网关ID --action connect --target 目标客户端ID
  发送消息: p2p -mode client --gateway 网关ID --action send --target 目标客户端ID --message "消息内容"
`

var (
	// P2p 命令定义了P2P连接工具的入口命令
	// 遵循GoFrame的Command对象定义规范，包含名称、用法、简短描述和执行函数
	P2p = gcmd.Command{
		// Name 为命令名称
		Name: "p2p",
		// Usage 描述命令的基本用法
		Usage: "p2p [options]",
		// Brief 提供命令的简短功能描述
		Brief: "P2P连接工具，支持网关和客户端模式，实现NAT穿透和点对点通信",
		// Description 提供命令的详细描述和使用帮助
		Description: p2pHelpDescription,
		// Func 为命令的执行函数，接收上下文和参数解析器
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()

			// 配置日志输出
			g.Log().SetConfigWithMap(g.Map{
				"level":  "all",
				"stdout": true,
			})

			parser, err = gcmd.Parse(g.MapStrBool{
				"m,mode":    true,
				"g,gateway": true,
				"a,action":  true,
				"t,target":  true,
				"m,message": true,
			})

			// 获取运行模式参数
			mode := parser.GetOpt("mode").String()

			// 根据不同模式调用服务层对应的方法
			switch mode {
			case "gateway":
				// 启动网关服务器模式
				g.Log().Debug(ctx, "开始执行gatway")
				s.Group("/ws", func(group *ghttp.RouterGroup) {
					group.Middleware(ghttp.MiddlewareHandlerResponse)
					service.P2P().GatewayStart(ctx, group)
				})
				s.Run()
			case "client":
				// 获取客户端模式所需的参数
				g.Log().Debug(ctx, "开始执行client")
				//addrs := []string{"/ip4/127.0.0.1/tcp/51888", "/ip4/192.168.50.173/tcp/51888"}
				//addr := "/ip4/192.168.50.173/tcp/51888/p2p/12D3KooWJKBB9bF9MjqgsFYUUsPBG249FDq7a3ZdaYc9iw8G78JQ"
				//addrs := "WyIvaXA0LzEyNy4wLjAuMS90Y3AvNTE4ODgiLCIvaXA0LzE5Mi4xNjguNTAuMTczL3RjcC81MTg4OCJd"
				wsStr := "ws://192.168.50.173:51888/ws"
				err = service.P2P().Start(ctx, wsStr)
			case "dht":

			default:
				// 显示帮助信息
				g.Log().Info(ctx, p2pHelpDescription)
			}

			return
		},
	}
)
