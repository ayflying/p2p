package v1

import "github.com/gogf/gf/v2/frame/g"

type UpdateReq struct {
	g.Meta  `path:"/system/update" tags:"system" method:"get" sm:"更新服务端"`
	Url     string `json:"url" dc:"更新地址"`
	Version string `json:"version" dc:"当前版本"`
}
type UpdateRes struct {
}
