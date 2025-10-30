package p2p

import (
	"fmt"
	"time"

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
	type tcpCfgType struct {
		Key       string `json:"key"`
		Ip        string `json:"ip"`
		Port      int    `json:"port"`
		LocalPort int    `json:"local_port"`
	}
	var tcpCfgList []*tcpCfgType
	tcpCfg, err := g.Cfg("proxy").Get(gctx.New(), "tcp")
	if err == nil {
		tcpCfg.Scan(&tcpCfgList)
		for _, v := range tcpCfgList {
			go s.Tcp(v.Key, v.Port, v.LocalPort, v.Ip)
			time.Sleep(2 * time.Second)
		}
	}

	time.Sleep(1 * time.Second)
	type httpCfgType struct {
		Key   string `json:"key"`
		Ip    string `json:"ip"`
		Port  int    `json:"port"`
		Cname string `json:"cname"`
	}
	var httpCfgList []*httpCfgType
	httpCfg, err := g.Cfg("proxy").Get(gctx.New(), "tcp")
	if err == nil {
		httpCfg.Scan(&httpCfgList)
		for _, v := range httpCfgList {
			go s.Http(v.Key, v.Cname, v.Port, v.Ip)
			time.Sleep(2 * time.Second)
		}
	}

}

var TcpList = make(map[int]*gtcp.Conn)

func (s *sP2P) TcpAck(port int) *gtcp.Conn {
	if v, ok := TcpList[port]; ok {
		return v
	} else {
		g.Log().Errorf(gctx.New(), "端口:%v不存在", port)
		return nil
	}
}

func (s *sP2P) Tcp(key string, toPort, myPort int, ip string) {
	if ip == "" {
		ip = "127.0.0.1"
	}

	// 建立p2p连接
	err := s.DiscoverAndConnect(key)
	if err != nil {
		g.Log().Errorf(gctx.New(), "连接失败:%v", err)
		time.Sleep(3 * time.Second)
		s.Tcp(key, toPort, myPort, ip)
		return
	}
	err = gtcp.NewServer(fmt.Sprintf("0.0.0.0:%d", myPort), func(conn *gtcp.Conn) {
		TcpList[toPort] = conn
		defer conn.Close()
		for {
			ctx := gctx.New()
			data, err := conn.Recv(-1)

			if len(data) > 0 {
				g.Log().Debugf(gctx.New(), "本地接口收到内容:%v", string(data))
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

func (s *sP2P) Http(key string, cname string, toPort int, ip string) {
	if toPort == 0 {
		toPort = 80
	}

	// 建立p2p连接
	err := s.DiscoverAndConnect(key)
	if err != nil {
		g.Log().Errorf(gctx.New(), "连接失败:%v", err)
		time.Sleep(3 * time.Second)
		s.Http(key, cname, toPort, ip)
		return
	}

}
