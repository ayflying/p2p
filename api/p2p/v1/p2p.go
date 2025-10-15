package v1

import "github.com/gogf/gf/v2/frame/g"

type ConnectReq struct {
	g.Meta   `path:"/p2p/connect" tags:"p2p" method:"get" sm:"连接到目标主机"`
	TargetID string `json:"id"`
}
type ConnectRes struct {
	g.Meta `mime:"text/html" example:"string"`
}

type SendReq struct {
	g.Meta   `path:"/p2p/send" tags:"p2p" method:"get" sm:"发送消息"`
	TargetID string `json:"id"`
	Data     string `json:"data"`
}
type SendRes struct {
	g.Meta `mime:"text/html" example:"string"`
}

type IpReq struct {
	g.Meta `path:"/p2p/ip" tags:"p2p" method:"get" sm:"获取当前主机的IP地址"`
}
type IpRes struct {
	g.Meta `mime:"text/html" example:"string"`
}
