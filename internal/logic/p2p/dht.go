package p2p

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gogf/gf/v2/crypto/gsha1"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/libp2p/go-libp2p/core/peerstore"

	//"github.com/ipfs/boxo/ipns"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
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
func (s *sP2P) DHTStart(h host.Host, bootstrapPeers []string) (err error) {
	ctx := gctx.New()

	//打印节点地址（供其他节点手动加入时使用）
	s.printNodeAddrs(h)

	if len(bootstrapPeers) == 0 {
		bootstrapPeers = []string{
			//"/ip4/192.168.50.243/tcp/23333/p2p/12D3KooWESZtrm6AfqhC3oj5FsAbcSmePwHFFip3F2MPExrxHxwy",
			//"/ip4/192.168.50.243/udp/23333/quic-v1/p2p/12D3KooWESZtrm6AfqhC3oj5FsAbcSmePwHFFip3F2MPExrxHxwy",
			//
			//"/ip4/192.168.50.173/tcp/23333/p2p/12D3KooWKgW8WxncYzZ2h5erMbK3GfLGhNHFapPvhUc1KVmdZeRg",
			//"/ip4/192.168.50.173/udp/23333/quic-v1/p2p/12D3KooWKgW8WxncYzZ2h5erMbK3GfLGhNHFapPvhUc1KVmdZeRg",

			//肖晓
			"/ip4/192.168.50.244/tcp/23333/p2p/12D3KooWFAt3hTi2SaYNty4gxxBnLRFxJidRDcf4k8HqCUZZRY1W",
			"/ip4/192.168.50.244/udp/23333/quic-v1/p2p/12D3KooWFAt3hTi2SaYNty4gxxBnLRFxJidRDcf4k8HqCUZZRY1W",

			//廖玉龙
			"/ip4/192.168.50.210/tcp/23333/p2p/12D3KooWM8eE3i2EWB2wFVGM1URusBPHJrEQJGxKfKgPdxEMm9hn",
			"/ip4/192.168.50.210/udp/23333/quic-v1/p2p/12D3KooWM8eE3i2EWB2wFVGM1URusBPHJrEQJGxKfKgPdxEMm9hn",
		}

	}

	s.dht.bootstrapPeers = bootstrapPeers

	// 2. 通过官方 Bootstrap 节点加入公共 DHT 网络（完全去中心化入口）
	s.dht.KadDHT, err = s.joinGlobalDHT(ctx, h)
	if err != nil {
		g.Log().Infof(ctx, "加入 DHT 网络失败: %v", err)
		g.Log().Info(ctx, "开启私有节点服务端等待中...")
		return
	}
	g.Log().Debug(ctx, "✅ 成功启动完全去中心化 DHT 网络")

	// 3. 定期打印路由表（观察节点自动发现效果）
	go s.printRoutingTable(ctx, s.dht.KadDHT, 60*time.Second)

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

	// 1. 检查key是否以 /ay/ 开头
	if !strings.HasPrefix(key, "/ay/") {
		return fmt.Errorf("拒绝存储：key必须以 /ay/ 开头，当前key为 %s", key)
	}

	g.Log().Debugf(gctx.New(), "外部数据进行保存:key: %s, value: %s", key, value)

	// 限制数据大小（防止超大数据占用资源）
	if len(value) > 1024*1024 { // 1MB上限
		return fmt.Errorf("数据超过1MB，拒绝存储")
	}

	return nil
}

// Select 简单返回第一个数据（不做版本选择）
func (v *NoOpValidator) Select(key string, values [][]byte) (int, error) {
	g.Log().Debugf(gctx.New(), "外部数据进行选择:key: %s, values: %v", key, values)
	return 0, nil
}

// 加入全球公共 DHT 网络（通过官方 Bootstrap 节点，实现完全去中心化）
func (s *sP2P) joinGlobalDHT(ctx context.Context, localHost host.Host) (*dht.IpfsDHT, error) {
	// 2. 基于Host创建IpfsDHT实例（关键步骤）
	// 注意：需指定模式（Full/Client），私有网络中Bootstrap节点用Full模式，普通节点用Client模式
	dhtOpts := []dht.Option{
		dht.Mode(dht.ModeClient), // 普通节点用Client模式（轻量）
		// dht.Mode(dht.ModeServer), // Bootstrap节点用Full模式（存储完整路由表）
	}

	// 创建 DHT 实例（ModeServer：作为完整节点参与存储和路由）
	kadDHT, err := dht.New(
		ctx,
		localHost,
		dhtOpts...,
	)
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
			// 带超时的连接，确保失败后能释放
			connCtx, connCancel := context.WithTimeout(ctx, 20*time.Second)
			err = localHost.Connect(connCtx, p)
			if err != nil {
				g.Log().Debugf(connCtx, "⚠️ 连接本地种子节点 %s 失败: %v\n", p.ID.ShortString(), err)
			} else {
				g.Log().Debugf(connCtx, "✅ 连接本地种子节点成功: %s\n", p.ID.ShortString())
			}
			connCancel()
			success = true
		}
		if !success {
			g.Log().Debugf(ctx, "所有本地种子节点连接失败")
		}
	}

	if !success {
		g.Log().Debug(ctx, "所有本地种子节点连接失败，私有网络启动失败")
		return nil, fmt.Errorf("所有本地种子节点连接失败") // 连接失败时终止DHT启动
	}

	//if !success {
	//	// 连接 libp2p 官方 Bootstrap 节点（仅作为初始入口）
	//	officialBootstrapPeers := dht.DefaultBootstrapPeers // 官方节点列表
	//	fmt.Println("正在连接官方 Bootstrap 节点（初始入口）...")
	//
	//	for _, addr := range officialBootstrapPeers {
	//		peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	//		if err != nil {
	//			fmt.Printf("⚠️  解析官方节点失败: %v\n", err)
	//			continue
	//		}
	//
	//		// 添加节点地址到本地地址簿
	//		localHost.Peerstore().AddAddrs(peerInfo.ID, peerInfo.Addrs, peerstore.PermanentAddrTTL)
	//
	//		// 尝试连接（超时 10 秒）
	//		connCtx, connCancel := context.WithTimeout(ctx, 20*time.Second)
	//		err = localHost.Connect(connCtx, *peerInfo)
	//		connCancel()
	//
	//		if err != nil {
	//			fmt.Printf("⚠️  连接官方节点 %s 失败: %v\n", peerInfo.ID.ShortString(), err)
	//			continue
	//		}
	//		fmt.Printf("✅ 连接官方节点成功: %s\n", peerInfo.ID.ShortString())
	//		success = true
	//	}
	//
	//	// 只要连接上至少一个官方节点，即可加入网络（后续会自动发现更多节点）
	//	if !success {
	//		return nil, fmt.Errorf("无法连接任何官方 Bootstrap 节点，无法加入网络")
	//	}
	//}

	//// 4. 执行Bootstrap，加入私有网络
	//bootstrapCfg := bootstrap.BootstrapConfig{
	//	BootstrapPeers: func() []peer.AddrInfo {
	//		seedPeers, _ := s.parseSeedNodes(s.dht.bootstrapPeers)
	//		return seedPeers
	//	}, // 私有Bootstrap节点列表
	//	//MinPeers:      1,              // 至少连接1个Bootstrap节点
	//	Period: 30 * time.Second, // 禁用定期重连
	//	//ConnectionMgr: connMgr,        // 关联连接管理器
	//}
	//if _, err = bootstrap.Bootstrap(localHost.ID(), localHost, kadDHT, bootstrapCfg); err != nil {
	//	return nil, fmt.Errorf("节点Bootstrap失败: %v", err)
	//}

	// 启动 DHT（自动发现其他节点，构建路由表，脱离对官方节点的依赖,带超时，避免阻塞）
	bootCtx, bootCancel := context.WithTimeout(ctx, 60*time.Second)
	err = kadDHT.Bootstrap(bootCtx)
	bootCancel()
	if err != nil {
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

	// 2. 带超时的存储，避免长期阻塞
	storeCtx, storeCancel := context.WithTimeout(ctx, 60*time.Second)
	defer storeCancel()

	g.Log().Debugf(storeCtx, "StoreToDHT key: %s, value: %s", key, value)
	err = s.dht.KadDHT.PutValue(storeCtx, key, []byte(value))
	if err != nil {
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
	localCtx, localCancel := context.WithTimeout(ctx, 10*time.Second)
	defer localCancel()
	localValue, err := s.dht.KadDHT.GetValue(localCtx, key)
	if err == nil {
		g.Log().Debugf(ctx, "✅ 本地查找成功（数据在当前节点）")
		return string(localValue), nil
	}
	g.Log().Debugf(ctx, "⚠️ 本地查找失败: %v，开始重试网络查找...", err)

	// 2. 多次重试网络查找
	for i := 0; i < maxRetries; i++ {
		findCtx, findCancel := context.WithTimeout(ctx, 60*time.Second)
		g.Log().Debugf(findCtx, "🔍 第%d次查找（共%d次）...", i+1, maxRetries)
		value, err2 := s.dht.KadDHT.GetValue(findCtx, key)
		findCancel()
		if err2 == nil {
			g.Log().Debugf(findCtx, "✅ 第%d次查找成功", i+1)
			return string(value), nil

		}
		g.Log().Debugf(ctx, "⚠️ 第%d次查找失败: %v，等待重试...", i+1, err2)
		time.Sleep(retryInterval)
	}

	return "", fmt.Errorf("超过最大重试次数（%d次），未找到数据", maxRetries)
}

// 定期打印路由表（观察节点自动发现情况）
func (s *sP2P) printRoutingTable(ctx context.Context, kadDHT *dht.IpfsDHT, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("路由表打印goroutine已退出")
			return
		case <-ticker.C:
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
}

// 打印节点地址（供其他节点手动加入时使用）
func (s *sP2P) printNodeAddrs(host host.Host) {
	fmt.Println("节点地址（公网地址将自动同步到DHT）:")
	for _, addr := range host.Addrs() {
		fullAddr := fmt.Sprintf("%s/p2p/%s", addr, host.ID())
		ipStr, _ := addr.ValueForProtocol(multiaddr.P_IP4)
		ipObj := net.ParseIP(ipStr)
		if ipObj.IsPrivate() || ipObj.IsLoopback() {
			fmt.Printf("%s\n", fullAddr)
		} else {
			fmt.Printf("%s\n", fullAddr)
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
