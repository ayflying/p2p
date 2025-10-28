package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strconv"
	"time"

	"github.com/ayflying/p2p/internal/service"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtcp"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/os/gtimer"
	"github.com/gogf/gf/v2/util/grand"
	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/multiformats/go-multiaddr"
)

// 客户端
type Client struct {
	ctx        context.Context
	Id         string
	gatewayURL string
	host       host.Host
	wsConn     *websocket.Conn    // WebSocket连接
	peers      map[string]peer.ID // 存储已连接的节点
	//tcp map[string]
}

type Message struct {
	Type string `json:"type" dc:"消息类型"`
	Port int    `json:"port,omitempty" dc:"请求端口"`
	Data []byte `json:"data" dc:"消息数据"`
	From string `json:"from" dc:"发送方ID"`
}

// SendP2P 发送格式化消息
func (s *sP2P) SendP2P(targetID string, typ string, data []byte) (err error) {
	if typ == "" {
		typ = "message"
	}
	message := &Message{
		Type: typ,
		From: s.client.Id,
		Data: data,
	}
	err = s.sendData(targetID, gjson.MustEncode(message))
	return
}

func (s *sP2P) Start(wsStr string) (err error) {
	var ctx = gctx.New()
	hostObj, err := s.CreateLibp2pHost(ctx, 0)
	if err != nil {
		g.Log().Error(ctx, err)
	}
	//defer hostObj.Close()

	// 创建客户端实例
	s.client = &Client{
		ctx:        ctx,
		Id:         hostObj.ID().String(),
		gatewayURL: wsStr,
		host:       hostObj,
		peers:      make(map[string]peer.ID),
	}

	// 设置流处理函数（处理P2P消息）
	hostObj.SetStreamHandler(protocol.ID(ProtocolID), s.handleStream)

	for {
		// 连接网关（WebSocket）
		if err = s.connectGateway(); err != nil {
			g.Log().Errorf(ctx, "连接网关失败,60秒后重试: %v", err)
			time.Sleep(60 * time.Second)
		} else {
			break
		}
	}

	// 启动网关消息接收协程
	go s.receiveGatewayMessages(ctx)

	//启动代理初始化
	s.ProxyInit()

	return
}

// 创建libp2p主机
func (s *sP2P) CreateLibp2pHost(ctx context.Context, port int) (host.Host, error) {
	if port == 0 {
		port = grand.N(50000, 55000)
		//port = 53533
	}

	// 配置监听地址
	//listenAddr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)
	var listenAddrs = []string{
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port),         // 随机 TCP 端口
		fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", port), // 随机 UDP 端口（QUIC 协议，提升打洞成功率）
	}

	// 1. 生成密钥对并初始化节点（确保身份有效）
	s.privKey, _ = s.generateFixedKey()

	// 3. 手动创建并挂载连接管理器（v0.43.0兼容）
	connMgr, err := connmgr.NewConnManager(
		100,                                     // LowWater：连接数低于此值时不主动断开
		500,                                     // HighWater：连接数高于此值时主动清理无效连接
		connmgr.WithGracePeriod(30*time.Second), // 宽限期：新连接30秒内不被清理
	)

	// 创建主机
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(listenAddrs...),
		libp2p.Identity(s.privKey),
		libp2p.EnableRelay(), // 启用中继兜底
		// 关键：通过WithConnManager选项注入连接管理器
		libp2p.ConnectionManager(connMgr),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.NATPortMap(), // 自动尝试路由器端口映射（跨网络必备）
	)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, fmt.Errorf("创建ConnManager失败: %v", err)
	}

	g.Log().Debugf(ctx, "当前p2p的分享地址：%v", h.Addrs())

	return h, err
}

// 连接网关（WebSocket）
func (s *sP2P) connectGateway() (err error) {
	var ctx = gctx.New()
	conn, _, err := websocket.DefaultDialer.Dial(s.client.gatewayURL, nil)
	if err != nil {
		gtimer.SetTimeout(ctx, 3*time.Minute, func(ctx context.Context) {
			err = s.connectGateway()
			return
		})
		return fmt.Errorf("WebSocket连接失败: %v", err)
	}
	//defer conn.Close()

	s.client.wsConn = conn
	g.Log().Infof(ctx, "已连接网关成功，客户端ID: %s", s.client.Id)

	// 注册到网关
	if err = s.register(); err != nil {
		g.Log().Fatalf(ctx, "注册到网关失败: %v", err)
	}

	g.Log().Infof(ctx, "已注册到网关，客户端ID: %s", s.client.Id)
	return
}

// 注册到网关
func (s *sP2P) register() error {
	selfAddrs := s.client.host.Peerstore().Addrs(s.client.host.ID())
	// 收集地址信息
	addrs := make([]string, len(selfAddrs))
	for i, addr := range selfAddrs {
		addrs[i] = addr.String()
	}

	// 构建注册消息
	msg := GatewayMessage{
		Type: MsgTypeRegister,
		From: s.client.Id,
		Data: gjson.MustEncode(g.Map{
			"peer_id": s.client.host.ID().String(),
			"addrs":   addrs,
		}),
	}

	return s.sendToGateway(msg)
}

// 发现并连接目标节点
func (s *sP2P) DiscoverAndConnect(targetID string) error {
	// 发送发现请求
	msg := GatewayMessage{
		Type: MsgTypeDiscover,
		From: s.client.Id, // 发送方是自己
		To:   "gateway",   // 接收方是网关
		Data: gjson.MustEncode(g.Map{
			"target_id": targetID,
		}),
	}
	if err := s.sendToGateway(msg); err != nil {
		return err
	}
	return nil
}

// 处理P2P流
func (s *sP2P) handleStream(stream network.Stream) {
	ctx := gctx.New()
	defer stream.Close()
	//var err error

	//peerID := stream.Conn().RemotePeer().String()
	//glog.Infof(ctx, "收到来自 %s 的连接", peerID)

	// 读取数据
	buf := make([]byte, 1024)
	var msg []byte

	for {
		n, err := stream.Read(buf)
		msg = append(msg, buf[:n]...)
		// 再判断错误
		if err != nil {
			if err == io.EOF {
				// EOF 是正常结束，不算错误
				err = nil
				break
			} else {
				return
			}
		}
		//if err != nil {
		//	glog.Errorf(ctx, "读取流数据失败: %v", err)
		//	return
		//}
	}

	// 解析消息
	var message *Message
	if err := gjson.DecodeTo(msg, &message); err != nil {
		g.Log().Debugf(ctx, "解析消息失败: %v", msg)
	}

	//g.Log().Debugf(ctx, "收到来自 %s 的消息: %v ", peerID, gjson.MustEncodeString(message))
	switch message.Type {
	case "proxy":
		var data *ProxyType
		gjson.DecodeTo(message.Data, &data)
		//g.Dump(data)
		// Client
		for {
			if conn, err := gtcp.NewConn(fmt.Sprintf("%s:%v", data.Ip, data.Port)); err == nil {
				if b, err := conn.SendRecv([]byte(gtime.Datetime()), -1); err == nil {
					fmt.Println(string(b), conn.LocalAddr(), conn.RemoteAddr())

					err = s.SendP2P(message.From, "proxy_ack", gjson.MustEncode(&ProxyType{
						Ip:   ip,
						Port: message.Port,
						Data: b,
					}))
				} else {
					fmt.Println(err)
				}
				conn.Close()
			} else {
				//glog.Error(err)
			}
			//time.Sleep(time.Second)
		}

		//conn, err := gtcp.NewConn(fmt.Sprintf("%s:%v", data.Ip, data.Port))
		//if err != nil {
		//	g.Log().Errorf(ctx, "连接失败:%v", err)
		//	return
		//}
		//defer conn.Close()

	}

}

// 发送数据到目标节点
func (s *sP2P) sendData(targetID string, data []byte) error {
	peerID, exists := s.client.peers[targetID]
	if !exists {
		return fmt.Errorf("未找到目标节点 %s 的连接", targetID)
	}

	// 创建流
	stream, err := s.client.host.NewStream(gctx.New(), peerID, protocol.ID(ProtocolID))
	if err != nil {
		return err
	}
	defer stream.Close()

	// 发送数据
	_, err = stream.Write(data)
	return err
}

// 处理网关的打洞请求
func (s *sP2P) handlePunchRequest(data json.RawMessage) error {
	ctx := gctx.New()
	var punchData struct {
		FromID string   `json:"from_id"`
		PeerID string   `json:"peer_id"`
		Addrs  []string `json:"addrs"`
	}

	if err := json.Unmarshal(data, &punchData); err != nil {
		return err
	}

	// 解析PeerID
	peerID, err := peer.Decode(punchData.PeerID)
	if err != nil {
		return err
	}

	// 解析地址
	addrs := make([]multiaddr.Multiaddr, len(punchData.Addrs))
	for i, addrStr := range punchData.Addrs {
		addr, err := multiaddr.NewMultiaddr(addrStr)
		if err != nil {
			glog.Errorf(ctx, "解析地址失败: %v", err)
			continue
		}
		addrs[i] = addr
	}

	// 添加到peerstore
	s.client.host.Peerstore().AddAddrs(peerID, addrs, peerstore.PermanentAddrTTL)

	// 立即尝试连接（关键：协调时机）
	glog.Infof(ctx, "收到打洞请求，尝试连接 %s", punchData.FromID)
	go func() {
		time.Sleep(500 * time.Millisecond) // 稍微延迟，确保双方都准备好
		if err := s.client.host.Connect(ctx, peer.AddrInfo{
			ID:    peerID,
			Addrs: addrs,
		}); err != nil {
			glog.Errorf(ctx, "打洞连接失败: %v", err)
			return
		}

		glog.Infof(ctx, "成功连接到 %s", punchData.FromID)
		s.client.peers[punchData.FromID] = peerID
	}()

	return nil
}

// 发送消息到网关
func (s *sP2P) sendToGateway(msg GatewayMessage) (err error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	err = s.client.wsConn.WriteMessage(websocket.TextMessage, data)
	return
}

// 接收网关消息
func (s *sP2P) receiveGatewayMessages(ctx context.Context) {
	for {
		_, data, err := s.client.wsConn.ReadMessage()
		if err != nil {
			g.Log().Errorf(ctx, "接收网关消息失败: %v", err)

			gtimer.SetTimeout(ctx, 10*time.Second, func(ctx context.Context) {
				err = s.connectGateway()
				return
			})
			return
		}

		var msg GatewayMessage
		if err = json.Unmarshal(data, &msg); err != nil {
			glog.Errorf(ctx, "解析网关消息失败: %v", err)
			continue
		}

		// 验证消息是否发给自己（to必须是当前客户端ID或空）
		if msg.To != "" && msg.To != s.client.Id {
			g.Log().Debugf(ctx, "忽略非本客户端的消息，from=%s, to=%s", msg.From, msg.To)
			continue
		}

		// 处理不同类型消息
		switch msg.Type {
		case MsgTypeRegisterAck:
			g.Log().Infof(ctx, "注册成功")

		case MsgTypeDiscoverAck:
			var msgData struct {
				Found    bool     `json:"found"`
				PeerID   string   `json:"peer_id,omitempty"`
				Addrs    []string `json:"addrs,omitempty"`
				TargetID string   `json:"target_id"`
			}
			if err = gjson.DecodeTo(msg.Data, &msgData); err != nil {
				g.Log().Errorf(ctx, "解析发现响应失败: %v", err)
				continue
			}

			if !msgData.Found {
				g.Log().Debug(ctx, "gateway未找到目标节点")
				continue
			}

			// 解析并连接目标节点
			peerID, err := peer.Decode(msgData.PeerID)
			if err != nil {
				glog.Errorf(ctx, "解析PeerID失败: %v", err)
				continue
			}

			g.Log().Infof(ctx, "准备开始打洞到目标节点：%v", msgData.TargetID)

			addrs := make([]multiaddr.Multiaddr, len(msgData.Addrs))
			for i, addrStr := range msgData.Addrs {
				addr, err := multiaddr.NewMultiaddr(addrStr)
				if err != nil {
					glog.Errorf(ctx, "解析地址失败: %v", err)
					continue
				}
				addrs[i] = addr
			}

			s.client.host.Peerstore().AddAddrs(peerID, addrs, peerstore.PermanentAddrTTL)

			// 立即尝试连接
			go func(targetPeerID peer.ID, targetID string) {
				time.Sleep(500 * time.Millisecond) // 协调时机
				if err = s.client.host.Connect(ctx, peer.AddrInfo{
					ID:    targetPeerID,
					Addrs: addrs,
				}); err != nil {
					g.Log().Errorf(ctx, "连接失败: %v", err)
					return
				}

				g.Log().Infof(ctx, "成功连接到目标节点：%v", targetID)
				s.client.peers[targetID] = targetPeerID
			}(peerID, msgData.TargetID)

		case MsgTypePunchRequest:
			err = s.handlePunchRequest(msg.Data)

		case MsgTypeError:
			var data struct {
				Error string `json:"error"`
			}
			json.Unmarshal(msg.Data, &data)
			glog.Errorf(ctx, "网关错误: %s", data.Error)
		case MsgUpdate: //更新节点信息
			var msgData struct {
				Files []struct {
					File []byte `json:"file"`
					Name string `json:"name"`
				} `json:"files"`
			}
			//var msgData *dataType
			json.Unmarshal(msg.Data, &msgData)
			for _, v := range msgData.Files {
				err = gfile.PutBytes(path.Join("download", v.Name), v.File)
			}
			// 更新器路径（假设与主程序同目录）
			//updaterPath := filepath.Join(filepath.Dir(selfPath), "updater.exe")

			g.Log().Info(ctx, "文件接收完成")
			// 开始覆盖文件与重启
			err = service.System().Update(ctx)

			//// 调用不同系统的更新服务
			//service.OS().Update(msgData.Version, msgData.Server)

		}
	}
}

// 提取libp2p节点的本地TCP监听端口
func (s *sP2P) getLocalTCPPorts(host host.Host) ([]int, error) {
	ports := make(map[int]struct{}) // 去重

	// 遍历所有本地监听地址
	for _, addr := range host.Addrs() {
		// 提取TCP端口
		portStr, err := addr.ValueForProtocol(multiaddr.P_TCP)
		if err != nil {
			continue // 跳过非TCP地址
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}

		ports[port] = struct{}{}
	}

	// 转换为切片返回
	result := make([]int, 0, len(ports))
	for port := range ports {
		result = append(result, port)
	}
	return result, nil
}
