package main

import (
	"fmt"
	"os"

	_ "github.com/ayflying/p2p/internal/logic"
	_ "github.com/ayflying/p2p/internal/packed"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	//步骤1：加载驱动
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"

	"github.com/ayflying/p2p/internal/cmd"
	"github.com/gogf/gf/v2/os/gctx"
)

var (
	ctx = gctx.GetInitCtx()
)

func main() {
	g.Log().Infof(ctx, "启动文件最后修改时间:%v", gtime.New(gfile.MTime(gcmd.GetArg(0).String())).String())
	//g.Dump("v1.0.0.2")

	if ok := gfile.Exists("runtime"); !ok {
		gfile.Mkdir("runtime")
	}

	//daili()

	cmd.Main.Run(ctx)
}

func daili() {

	// 读取HTTP代理环境变量（小写/大写通常都兼容，部分系统可能用大写）
	httpProxy := os.Getenv("http_proxy")
	if httpProxy == "" {
		httpProxy = os.Getenv("HTTP_PROXY")
	}

	// 读取HTTPS代理环境变量
	httpsProxy := os.Getenv("https_proxy")
	if httpsProxy == "" {
		httpsProxy = os.Getenv("HTTPS_PROXY")
	}

	// 读取无代理列表（不使用代理的域名/IP）
	noProxy := os.Getenv("no_proxy")
	if noProxy == "" {
		noProxy = os.Getenv("NO_PROXY")
	}

	fmt.Printf("HTTP 代理: %s\n", httpProxy)
	fmt.Printf("HTTPS 代理: %s\n", httpsProxy)
	fmt.Printf("无代理列表: %s\n", noProxy)
}
