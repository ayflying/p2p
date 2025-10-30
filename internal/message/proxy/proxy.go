package proxy

import (
	"fmt"

	"github.com/ayflying/p2p/api/p2p/v1"
	"github.com/ayflying/p2p/internal/service"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtcp"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gtime"
)

var New = Proxy{}

var (
//ip, _ = service.P2P().GetIPv4PublicIP()
)

type Proxy struct {
	Ip   string `json:"ip"`
	Port int    `json:"port"`
	Data []byte `json:"data"`
}

func (p Proxy) Message(msg *v1.Message) (err error) {
	var data *Proxy

	gjson.DecodeTo(msg.Data, &data)
	//g.Dump(data)
	// Client
	go func() {
		if conn, err := gtcp.NewConn(fmt.Sprintf("%s:%v", data.Ip, data.Port)); err == nil {
			defer conn.Close()

			err = conn.Send(data.Data)
			if b, err := conn.SendRecv([]byte(gtime.Datetime()), -1); err == nil {
				//fmt.Println(string(b), conn.LocalAddr(), conn.RemoteAddr())

				err = service.P2P().SendP2P(msg.From, "proxy_ack", gjson.MustEncode(&Proxy{
					Ip:   data.Ip,
					Port: data.Port,
					Data: b,
				}))
				if err != nil {
					g.Log().Errorf(gctx.New(), "发送ACK失败:%v", err)
				}

			} else {
				fmt.Println(err)
			}

		} else {
			//glog.Error(err)
		}
		//time.Sleep(time.Second)

	}()

	//conn, err := gtcp.NewConn(fmt.Sprintf("%s:%v", data.Ip, data.Port))
	//if err != nil {
	//	g.Log().Errorf(ctx, "连接失败:%v", err)
	//	return
	//}
	//defer conn.Close()
	return
}
