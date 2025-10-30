package system

import (
	"github.com/ayflying/p2p/internal/service"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

type sSystem struct{}

func New() *sSystem {
	return &sSystem{}
}

func init() {
	service.RegisterSystem(New())

	getDev, _ := g.Cfg().GetWithEnv(gctx.New(), "dev")
	if !getDev.Bool() {
		err := service.System().CheckUpdate()
		if err != nil {
			g.Log().Errorf(gctx.New(), "检查更新失败：%v", err)
		}
	} else {
		g.Log().Debugf(gctx.New(), "开发模式，不检查更新")
	}
}

func (s *sSystem) Init() {}
