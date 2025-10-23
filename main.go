package main

import (
	_ "github.com/ayflying/p2p/internal/logic"
	_ "github.com/ayflying/p2p/internal/packed"
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
	if ok := gfile.Exists("runtime"); !ok {
		gfile.Mkdir("runtime")
	}

	cmd.Main.Run(ctx)
}
