package v1

import "github.com/gogf/gf/v2/frame/g"

type ConnectReq struct {
	g.Meta   `path:"/message/connect" tags:"message" method:"get" sm:"连接到目标主机"`
	TargetID string `json:"id"`
}
type ConnectRes struct {
	g.Meta `mime:"text/html" example:"string"`
}

type SendReq struct {
	g.Meta   `path:"/message/send" tags:"message" method:"get" sm:"发送消息"`
	TargetID string `json:"id"`
	Data     string `json:"data"`
}
type SendRes struct {
	g.Meta `mime:"text/html" example:"string"`
}

type IpReq struct {
	g.Meta `path:"/message/ip" tags:"message" method:"get" sm:"获取当前主机的IP地址"`
}
type IpRes struct {
	g.Meta `mime:"text/html" example:"string"`
}

type Message struct {
	Type string `json:"type" dc:"消息类型"`
	//Port int    `json:"port,omitempty" dc:"请求端口"`
	Data []byte `json:"data" dc:"消息数据"`
	From string `json:"from" dc:"发送方ID"`
}
