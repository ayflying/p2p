package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/ayflying/p2p/internal/service"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
)

var (
	// DHT 命令定义了P2P连接工具的入口命令
	// 遵循GoFrame的Command对象定义规范，包含名称、用法、简短描述和执行函数
	DHT = gcmd.Command{
		// Name 为命令名称
		Name: "dht",
		// Usage 描述命令的基本用法
		Usage: "dht [options]",
		// Brief 提供命令的简短功能描述
		Brief: "P2P连接工具，支持网关和客户端模式，实现NAT穿透和点对点通信",
		// Description 提供命令的详细描述和使用帮助
		Description: p2pHelpDescription,
		// Func 为命令的执行函数，接收上下文和参数解析器
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			g.Log().Debug(ctx, "开始执行dht")

			parser, err = gcmd.Parse(g.MapStrBool{
				"p,port": true,
			})
			port := parser.GetOpt("port", "23333").Int()

			h, _ := service.P2P().CreateLibp2pHost(ctx, port)
			err = service.P2P().DHTStart(h, nil)
			if err != nil {
				g.Log().Error(ctx, err)
			}

			go func() {
				time.Sleep(5 * time.Second)
				publicIp, _ := service.P2P().GetIPv4PublicIP()
				validKey := fmt.Sprintf("%v/ip", h.ID())
				dataValue := fmt.Sprintf("来自节点 %s 的数据:%v", h.ID().ShortString(), publicIp)
				if err = service.P2P().StoreToDHT(ctx, validKey, dataValue); err != nil {
					fmt.Printf("❌ 存储失败: %v\n", err)
				} else {
					fmt.Printf("✅ 存储成功\nKey: %s\nValue: %s\n", validKey, dataValue)
				}
			}()

			s.Run()
			return
		},
	}
)
