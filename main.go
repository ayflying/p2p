package main

import (
	"github.com/ayflying/p2p/internal/consts"
	_ "github.com/ayflying/p2p/internal/logic"
	_ "github.com/ayflying/p2p/internal/packed"
	"github.com/ayflying/p2p/internal/service"
	"github.com/gogf/gf/v2/os/gfile"

	//步骤1：加载驱动
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"

	"github.com/ayflying/p2p/internal/cmd"
	"github.com/gogf/gf/v2/os/gctx"
)

var (
	ctx = gctx.GetInitCtx()
)

func main() {
	// 启动系统托盘
	service.OS().Load(consts.Name, consts.Name+"服务端", "manifest/images/favicon.ico")

	if ok := gfile.Exists("runtime"); !ok {
		gfile.Mkdir("runtime")
	}

	cmd.Main.Run(ctx)
}
