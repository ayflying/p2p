package p2p

import (
	"context"
	"fmt"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
)

// 初始化无服务器DHT（作为节点加入DHT网络）
func (s *sP2P) DHTStart(ctx context.Context, h host.Host) (*dht.IpfsDHT, error) {
	// 创建DHT实例，设置为“客户端+服务端模式”（既可以查找数据，也可以存储数据）
	kdht, err := dht.New(ctx, h, dht.Mode(dht.ModeServer))
	if err != nil {
		return nil, fmt.Errorf("初始化DHT失败: %v", err)
	}

	// 启动DHT并加入网络（会自动发现网络中的其他DHT节点）
	if err := kdht.Bootstrap(ctx); err != nil {
		return nil, fmt.Errorf("DHT加入网络失败: %v", err)
	}

	fmt.Println("DHT初始化成功，节点ID:", h.ID().ShortString())
	return kdht, nil
}
