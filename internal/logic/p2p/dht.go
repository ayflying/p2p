package p2p

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/gogf/gf/v2/crypto/gsha1"
	"github.com/gogf/gf/v2/frame/g"

	//"github.com/ipfs/boxo/ipns"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/multiformats/go-multiaddr"
)

type DHTType struct {
	KadDHT         *dht.IpfsDHT
	bootstrapPeers []string
}

var (
//bootstrapPeers = []string{
//	"/ip4/192.168.50.173/tcp/53486/p2p/12D3KooWE3v9623SLukT9dKUQLjqAJrPvzoyRjoUh5MAVGDg69Rw",
//	"/ip4/192.168.50.173/udp/53486/quic-v1/p2p/12D3KooWE3v9623SLukT9dKUQLjqAJrPvzoyRjoUh5MAVGDg69Rw",
//}
)

// 初始化无服务器DHT（作为节点加入DHT网络）
func (s *sP2P) DHTStart(ctx context.Context, h host.Host, bootstrapPeers []string) (err error) {
	//打印节点地址（供其他节点手动加入时使用）
	s.printNodeAddrs(h)
	s.dht.bootstrapPeers = bootstrapPeers

	// 2. 通过官方 Bootstrap 节点加入公共 DHT 网络（完全去中心化入口）
	s.dht.KadDHT, err = s.joinGlobalDHT(ctx, h)
	if err != nil {
		log.Fatalf("加入 DHT 网络失败: %v", err)
	}
	fmt.Println("✅ 成功加入完全去中心化 DHT 网络")

	// 3. 定期打印路由表（观察节点自动发现效果）
	go s.printRoutingTable(s.dht.KadDHT, 60*time.Second)

	return
}

// 生成符合DHT规范的字符串Key
func (s *sP2P) generateStringDHTKey(str string) string {
	return gsha1.Encrypt(str)
	//fullKey := fmt.Sprintf("%s/%s", ProtocolID, str)
	//hash, _ := multihash.Sum([]byte(fullKey), multihash.SHA2_256, -1)
	//return ipns.key
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

// 加入全球公共 DHT 网络（通过官方 Bootstrap 节点，实现完全去中心化）
func (s *sP2P) joinGlobalDHT(ctx context.Context, localHost host.Host) (*dht.IpfsDHT, error) {
	// 创建 DHT 实例（ModeServer：作为完整节点参与存储和路由）
	kadDHT, err := dht.New(ctx, localHost, dht.Mode(dht.ModeServer))
	if err != nil {
		return nil, err
	}
	kadDHT.Validator = &NoOpValidator{}
	success := false

	if len(s.dht.bootstrapPeers) > 0 {
		fmt.Println("正在连接本地种子节点...")
		seedPeers, _ := s.parseSeedNodes(s.dht.bootstrapPeers)
		for _, p := range seedPeers {
			localHost.Peerstore().AddAddrs(p.ID, p.Addrs, peerstore.PermanentAddrTTL)
			if err = localHost.Connect(ctx, p); err != nil {
				fmt.Printf("⚠️  连接本地种子节点 %s 失败: %v\n", p.ID.ShortString(), err)
			} else {
				fmt.Printf("✅ 连接本地种子节点成功: %s\n", p.ID.ShortString())
			}
			if err != nil {
				fmt.Printf("⚠️  连接私有节点 %s 失败: %v\n", p.ID.ShortString(), err)
				continue
			}
			fmt.Printf("✅ 连接本地种子节点成功: %s\n", p.ID.ShortString())
			success = true
		}
		if !success {
			return nil, fmt.Errorf("所有本地种子节点连接失败")
		}
	} else {

		// 连接 libp2p 官方 Bootstrap 节点（仅作为初始入口）
		officialBootstrapPeers := dht.DefaultBootstrapPeers // 官方节点列表
		fmt.Println("正在连接官方 Bootstrap 节点（初始入口）...")

		for _, addr := range officialBootstrapPeers {
			peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
			if err != nil {
				fmt.Printf("⚠️  解析官方节点失败: %v\n", err)
				continue
			}

			// 添加节点地址到本地地址簿
			localHost.Peerstore().AddAddrs(peerInfo.ID, peerInfo.Addrs, peerstore.PermanentAddrTTL)

			// 尝试连接（超时 10 秒）
			connCtx, connCancel := context.WithTimeout(ctx, 10*time.Second)
			err = localHost.Connect(connCtx, *peerInfo)
			connCancel()

			if err != nil {
				fmt.Printf("⚠️  连接官方节点 %s 失败: %v\n", peerInfo.ID.ShortString(), err)
				continue
			}
			fmt.Printf("✅ 连接官方节点成功: %s\n", peerInfo.ID.ShortString())
			success = true
		}

		// 只要连接上至少一个官方节点，即可加入网络（后续会自动发现更多节点）
		if !success {
			return nil, fmt.Errorf("无法连接任何官方 Bootstrap 节点，无法加入网络")
		}
	}

	// 启动 DHT（自动发现其他节点，构建路由表，脱离对官方节点的依赖）
	if err = kadDHT.Bootstrap(ctx); err != nil {
		return nil, fmt.Errorf("DHT 初始化失败: %v", err)
	}

	// 等待路由表初步填充（新增：给路由表留出初始化时间）
	time.Sleep(5 * time.Second)
	return kadDHT, nil
}

// StoreToDHT 存储数据到 DHT（自动分布式存储）
func (s *sP2P) StoreToDHT(ctx context.Context, key string, value string) (err error) {
	key = s.generateStringDHTKey(key)
	key = fmt.Sprintf("%s/%s", ProtocolID, key)
	g.Log().Debugf(ctx, "StoreToDHT key: %s, value: %s", key, value)

	// 存储到本地
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	if err = s.dht.KadDHT.PutValue(ctx, key, []byte(value)); err != nil {
		return fmt.Errorf("本地存储失败: %v", err)
	}

	return
}

// FindFromDHT 从 DHT 查找数据（从网络节点获取）
func (s *sP2P) FindFromDHT(ctx context.Context, key string) (string, error) {
	maxRetries := 10                  // 最多重试5次
	retryInterval := 60 * time.Second // 每次重试间隔2秒（本地网络快）

	key = s.generateStringDHTKey(key)
	key = fmt.Sprintf("%s/%s", ProtocolID, key)
	g.Log().Debugf(ctx, "FindFromDHT key: %s", key)

	// 1. 先检查本地是否存储了数据（本地节点可能已保存）
	localValue, err := s.dht.KadDHT.GetValue(ctx, key)
	if err == nil {
		g.Log().Debugf(ctx, "✅ 本地查找成功（数据在当前节点）")
		return string(localValue), nil
	}
	g.Log().Debugf(ctx, "⚠️  本地查找失败: %v，开始重试网络查找...", err)

	// 2. 多次重试网络查找
	for i := 0; i < maxRetries; i++ {
		ctx2, cancel := context.WithTimeout(ctx, 120*time.Second) // 本地测试超时短一些
		defer cancel()

		g.Log().Debugf(ctx2, "🔍 第%d次查找（共%d次）...", i+1, maxRetries)
		value, err := s.dht.KadDHT.GetValue(ctx2, key)
		if err == nil {
			g.Log().Debugf(ctx2, "✅ 第%d次查找成功", i+1)
			return string(value), nil
		}
		g.Log().Debugf(ctx2, "⚠️  第%d次查找失败: %v，等待重试...", i+1, err)
		time.Sleep(retryInterval)
	}

	return "", fmt.Errorf("超过最大重试次数（%d次），未找到数据", maxRetries)
}

// 定期打印路由表（观察节点自动发现情况）
func (s *sP2P) printRoutingTable(kadDHT *dht.IpfsDHT, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		<-ticker.C
		peers := kadDHT.RoutingTable().ListPeers()
		fmt.Printf("\n📊 当前路由表节点数: %d（完全去中心化网络节点）\n", len(peers))
		if len(peers) > 0 {
			fmt.Println("前 5 个节点 ID:")
			for i, p := range peers[:min(5, len(peers))] {
				fmt.Printf("  %d. %s\n", i+1, p.ShortString())
			}
		}
	}
}

//// 定期打印节点状态（公网地址+路由表）
//func (s *sP2P) printStatus(interval time.Duration) {
//	ticker := time.NewTicker(interval)
//	for {
//		<-ticker.C
//		//publicIp, err := service.P2P().GetIPv4PublicIP()
//		//publicAddrs := s.getPublicAddrs()
//		peers := s.dht.KadDHT.RoutingTable().ListPeers()
//		fmt.Printf("\n===== 节点状态 =====")
//		fmt.Printf("\n公网地址数: %d（0表示穿透失败）\n", len(publicAddrs))
//		fmt.Printf("路由表节点数: %d（自动扩散结果）\n", len(peers))
//		fmt.Println("====================")
//	}
//}

// 打印节点地址（供其他节点手动加入时使用）
func (s *sP2P) printNodeAddrs(host host.Host) {
	fmt.Println("节点地址（公网地址将自动同步到DHT）:")
	for _, addr := range host.Addrs() {
		fullAddr := fmt.Sprintf("%s/p2p/%s", addr, host.ID())
		ipStr, _ := addr.ValueForProtocol(multiaddr.P_IP4)
		ip := net.ParseIP(ipStr)
		if ip.IsPrivate() || ip.IsLoopback() {
			fmt.Printf("  [内网] %s\n", fullAddr)
		} else {
			fmt.Printf("  [公网] %s\n", fullAddr)
		}
	}
}

func (s *sP2P) parseSeedNodes(seedAddrs []string) ([]peer.AddrInfo, error) {
	peers := make([]peer.AddrInfo, 0, len(seedAddrs))
	for _, addrStr := range seedAddrs {
		addr, err := multiaddr.NewMultiaddr(addrStr)
		if err != nil {
			return nil, err
		}
		peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			return nil, err
		}
		peers = append(peers, *peerInfo)
	}
	return peers, nil
}
