// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/gtcp"
	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p/core/host"
)

type (
	IP2P interface {
		// SendP2P 发送格式化消息
		SendP2P(targetID string, typ string, data []byte) (err error)
		Start(wsStr string) (err error)
		// 创建libp2p主机
		CreateLibp2pHost(ctx context.Context, port int) (host.Host, error)
		// 发现并连接目标节点
		DiscoverAndConnect(targetID string) error
		// 初始化无服务器DHT（作为节点加入DHT网络）
		DHTStart(h host.Host, bootstrapPeers []string) (err error)
		// StoreToDHT 存储数据到 DHT（自动分布式存储）
		StoreToDHT(ctx context.Context, key string, value string) (err error)
		// FindFromDHT 从 DHT 查找数据（从网络节点获取）
		FindFromDHT(ctx context.Context, key string) (string, error)
		GatewayStart(ctx context.Context, group *ghttp.RouterGroup) (err error)
		// 发送错误消息
		SendError(conn *websocket.Conn, errMsg string)
		// SendAll 发送消息给所有客户端
		SendAll(typ string, data any) (err error)
		// Send 发送消息给指定客户端
		Send(conn *websocket.Conn, typ string, data any) (err error)
		// 只获取IPv4公网IP（过滤IPv6结果）
		GetIPv4PublicIP() (string, error)
		ProxyInit()
		TcpAck(port int) *gtcp.Conn
		Tcp(key string, toPort int, myPort int, ip string)
		Http(key string, cname string, toPort int, ip string)
	}
)

var (
	localP2P IP2P
)

func P2P() IP2P {
	if localP2P == nil {
		panic("implement not found for interface IP2P, forgot register?")
	}
	return localP2P
}

func RegisterP2P(i IP2P) {
	localP2P = i
}
