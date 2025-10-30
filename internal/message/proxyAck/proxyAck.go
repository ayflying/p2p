package proxyAck

import (
	v1 "github.com/ayflying/p2p/api/p2p/v1"
	"github.com/ayflying/p2p/internal/service"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

type ProxyAck struct {
	Ip   string `json:"ip"`
	Port int    `json:"port"`
	Data []byte `json:"data"`
}

var New = ProxyAck{}

func (p ProxyAck) Message(msg *v1.Message) (err error) {
	var data *ProxyAck
	gjson.DecodeTo(msg.Data, &data)
	//g.Dump(data)

	g.Log().Debugf(gctx.New(), "收到ACK发送到端口:%v", data.Port)
	err = service.P2P().TcpAck(data.Port).Send(data.Data)

	return
}
