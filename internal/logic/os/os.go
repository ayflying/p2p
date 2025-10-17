package os

import (
	"github.com/ayflying/p2p/internal/service"
	"github.com/gogf/gf/v2/os/gcmd"
)

type systrayType struct {
	Icon    string `json:"icon" dc:"图标"`
	Title   string `json:"title" dc:"标题"`
	Tooltip string `json:"tooltip" dc:"提示"`
}

type sOS struct {
	systray *systrayType
}

func New() *sOS {
	return &sOS{
		systray: &systrayType{},
	}
}
func init() {
	service.RegisterOS(New())
}

func (s *sOS) Load(title string, tooltip string, ico string) {
	if title == "" {
		title = gcmd.GetArg(0).String()
	}
	if tooltip == "" {
		tooltip = gcmd.GetArg(0).String()
	}

	s.systray = &systrayType{
		Icon:    ico,
		Title:   title,
		Tooltip: tooltip,
	}
	s.start()
}

func (s *sOS) Update(version, server string) {
	s.update(version, server)
}
