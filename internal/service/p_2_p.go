// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/libp2p/go-libp2p/core/host"
)

type (
	IP2P interface {
		// SendP2P 发送格式化消息
		SendP2P(targetID string, typ string, data []byte) (err error)
		Start(ctx context.Context, wsStr string) (err error)
		// 创建libp2p主机
		CreateLibp2pHost(ctx context.Context, port int) (host.Host, error)
		// 发现并连接目标节点
		DiscoverAndConnect(targetID string) error
		// 初始化无服务器DHT（作为节点加入DHT网络）
		DHTStart(ctx context.Context, h host.Host) (err error)
		// 存储数据到DHT（比如存储“目标节点ID-公网地址”的映射）
		StoreAddrToDHT(ctx context.Context, key string, addr string) (err error)
		// 从DHT查找数据（比如根据节点ID查找其公网地址）
		FindAddrFromDHT(ctx context.Context, key string) (string, error)
		GatewayStart(ctx context.Context, group *ghttp.RouterGroup) (err error)
		// 只获取IPv4公网IP（过滤IPv6结果）
		GetIPv4PublicIP() (string, error)
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
