package p2p

import (
	"fmt"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtcp"
	"github.com/gogf/gf/v2/os/gctx"
)

type ProxyType struct {
	Ip   string `json:"ip"`
	Port int    `json:"port"`
	Data []byte `json:"data"`
}

func (s *sP2P) ProxyInit() {
	type cfgType struct {
		Key       string `json:"key"`
		Ip        string `json:"ip"`
		Port      int    `json:"port"`
		LocalPort int    `json:"local_port"`
	}
	var cfgList []*cfgType
	proxyCfg, err := g.Cfg("proxy").Get(gctx.New(), "tcp")
	if err == nil {
		proxyCfg.Scan(&cfgList)
		for _, v := range cfgList {
			go s.Tcp(v.Key, v.Port, v.LocalPort, v.Ip)

		}
	}

}

func (s *sP2P) Tcp(key string, toPort, myPort int, ip string) {
	if ip == "" {
		ip = "127.0.0.1"
	}

	// 建立p2p连接
	err := s.DiscoverAndConnect(key)
	if err != nil {

	}
	err = gtcp.NewServer(fmt.Sprintf("0.0.0.0:%d", myPort), func(conn *gtcp.Conn) {
		defer conn.Close()
		for {
			ctx := gctx.New()
			data, err := conn.Recv(-1)

			if len(data) > 0 {
				g.Log().Debugf(gctx.New(), "内容:%v", string(data))
				err = s.SendP2P(key, "proxy", gjson.MustEncode(&ProxyType{
					Ip:   ip,
					Port: toPort,
					Data: data,
				}))
				if err != nil {
					g.Log().Errorf(ctx, "发送失败:%v", err)
					s.Tcp(key, toPort, myPort, ip)
					return
				}
				//if err = conn.Send(append([]byte("> "), data...)); err != nil {
				//	fmt.Println(err)
				//}
			}
			if err != nil {
				break
			}
		}
	}).Run()
	if err != nil {
		g.Log().Error(gctx.New(), err)
	}

}
