package boot

import (
	"context"

	updateGithub "github.com/ayflying/update-github-release"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcron"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
)

func Boot() {

	getDev, _ := g.Cfg().GetWithEnv(gctx.New(), "dev")
	if !getDev.Bool() {
		var update = updateGithub.New("https://api.github.com/repos/ayflying/p2p/releases/latest")
		gfile.PutContents("version.txt", "v0.0.9")

		// 每天0点检查更新
		gcron.Add(gctx.New(), "0 0 0 * * *", func(ctx context.Context) {
			err := update.CheckUpdate(true)
			if err != nil {
				g.Log().Errorf(ctx, "检查更新失败：%v", err)
			}
		})

		go func() {
			//在协程中检查更新，预防主程序阻塞
			err := update.CheckUpdate(true)
			if err != nil {
				g.Log().Errorf(gctx.New(), "检查更新失败：%v", err)
			}
		}()
	} else {
		g.Log().Debugf(gctx.New(), "开发模式，不检查更新")
	}

}
