package p2p

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	//"github.com/ipfs/boxo/ipns"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
)

type DHTType struct {
	KadDHT *dht.IpfsDHT
}

// 初始化无服务器DHT（作为节点加入DHT网络）
func (s *sP2P) DHTStart(ctx context.Context, h host.Host) (err error) {
	//  创建自定义 DHT 选项，配置验证器
	dhtOpts := []dht.Option{
		//设置为“客户端+服务端模式”（既可以查找数据，也可以存储数据）
		dht.Mode(dht.ModeServer),
	}

	// 创建DHT实例，
	s.dht.KadDHT, err = dht.New(
		ctx,
		h,
		dhtOpts...,
		//dht.Mode(dht.ModeServer),
	)
	if err != nil {
		err = fmt.Errorf("初始化DHT失败: %v", err)
		return
	}

	// 关键：直接替换 DHT 实例的验证器
	// v0.35.1 版本中，IpfsDHT 结构体的 Validator 字段是公开可修改的
	s.dht.KadDHT.Validator = &NoOpValidator{}

	// 连接到DHT bootstrap节点（种子节点，帮助加入网络）
	// 这里使用libp2p官方的公共bootstrap节点，生产环境可替换为自己的节点
	bootstrapPeers := dht.DefaultBootstrapPeers
	for _, addr := range bootstrapPeers {
		peerInfo, _ := peer.AddrInfoFromP2pAddr(addr)
		h.Peerstore().AddAddrs(peerInfo.ID, peerInfo.Addrs, peerstore.PermanentAddrTTL)
		if err = h.Connect(ctx, *peerInfo); err != nil {
			fmt.Printf("连接bootstrap节点 %s 失败: %v\n", peerInfo.ID, err)
		} else {
			fmt.Printf("已连接bootstrap节点: %s\n", peerInfo.ID)
		}

	}

	// 启动DHT
	if err = s.dht.KadDHT.Bootstrap(ctx); err != nil {
		return
	}

	return
}

// 存储数据到DHT（比如存储“目标节点ID-公网地址”的映射）
func (s *sP2P) StoreAddrToDHT(ctx context.Context, key string, addr string) (err error) {
	// Key：目标节点ID（作为哈希键），Value：公网地址（需转成二进制）
	//key = "/ipns/" + key
	//key = s.generateStringDHTKey(key)
	key = fmt.Sprintf("%s/%s", ProtocolID, key)
	value := []byte(addr)

	// 存储数据（DHT会自动找到负责存储该Key的节点，并同步数据）
	if err = s.dht.KadDHT.PutValue(ctx, key, value); err != nil {
		return fmt.Errorf("key=%s,存储地址到DHT失败: %v", key, err)
	}

	g.Log().Info(ctx, "成功存储地址到DHT，Key=%s, Value=%s", key, addr)
	return
}

// 从DHT查找数据（比如根据节点ID查找其公网地址）
func (s *sP2P) FindAddrFromDHT(ctx context.Context, key string) (string, error) {
	// 查找数据（DHT会通过路由表层层跳转，找到负责存储该Key的节点并获取数据）
	//key = s.generateStringDHTKey(key)
	//key = "/ipns/" + key
	key = fmt.Sprintf("%s/%s", ProtocolID, key)
	g.Log().Debugf(ctx, "从DHT查找地址中...，Key=%s", key)
	value, err := s.dht.KadDHT.GetValue(ctx, key)
	if err != nil {
		return "", fmt.Errorf("从DHT查找地址失败: %v", err)
	}

	addr := string(value)
	fmt.Printf("从DHT找到地址，Key=%s, Value=%s\n", key, addr)
	return addr, nil
}

// 生成符合DHT规范的字符串Key
func (s *sP2P) generateStringDHTKey(str string) string {
	return ""
}

// 自定义验证器：不做任何校验，接受所有数据
type NoOpValidator struct{}

// Validate 总是返回成功，允许任何数据
func (v *NoOpValidator) Validate(key string, value []byte) error {
	return nil
}

// Select 简单返回第一个数据（不做版本选择）
func (v *NoOpValidator) Select(key string, values [][]byte) (int, error) {
	return 0, nil
}
