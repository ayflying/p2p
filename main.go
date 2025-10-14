package main

import (
	_ "github.com/ayflying/p2p/internal/packed"

	_ "github.com/ayflying/p2p/internal/logic"

	//步骤1：加载驱动
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"

	"github.com/ayflying/p2p/internal/cmd"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
