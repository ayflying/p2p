package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
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
}

func (s *sP2P) Start(ctx context.Context, wsStr string) (err error) {

	hostObj, err := s.createLibp2pHost(ctx, 0)
	if err != nil {
		g.Log().Error(ctx, err)
	}
	//defer hostObj.Close()

	// 创建客户端实例
	s.client = &Client{
		ctx:        ctx,
		Id:         uuid.New().String(),
		gatewayURL: wsStr,
		host:       hostObj,
		peers:      make(map[string]peer.ID),
	}

	// 设置流处理函数（处理P2P消息）
	hostObj.SetStreamHandler(ProtocolID, s.handleStream)

	// 连接网关（WebSocket）
	if err = s.connectGateway(); err != nil {
		glog.Fatalf(ctx, "连接网关失败: %v", err)
	}

	// 注册到网关
	if err = s.register(); err != nil {
		glog.Fatalf(ctx, "注册到网关失败: %v", err)
	}

	// 启动网关消息接收协程
	go s.receiveGatewayMessages()

	g.Log().Infof(ctx, "已连接网关成功，客户端ID: %s", s.client.Id)
	//g.Log().Infof(ctx,"当前地址：http://127.0.0.1/")

	//select {
	//case <-ctx.Done():
	//}
	return
}

// 创建libp2p主机
func (s *sP2P) createLibp2pHost(ctx context.Context, port int) (host.Host, error) {
	// 配置监听地址
	listenAddr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)

	// 创建主机
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(listenAddr),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
	)
	return h, err
}

// 连接网关（WebSocket）
func (s *sP2P) connectGateway() (err error) {
	conn, _, err := websocket.DefaultDialer.Dial(s.client.gatewayURL, nil)
	if err != nil {
		return fmt.Errorf("WebSocket连接失败: %v", err)
	}
	//defer conn.Close()

	s.client.wsConn = conn
	return
}

// 注册到网关
func (s *sP2P) register() error {
	// 收集地址信息
	addrs := make([]string, len(s.client.host.Addrs()))
	for i, addr := range s.client.host.Addrs() {
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
	defer stream.Close()

	peerID := stream.Conn().RemotePeer().String()
	glog.Infof(ctx, "收到来自 %s 的连接", peerID)

	// 读取数据
	buf := make([]byte, 1024)
	n, err := stream.Read(buf)
	if err != nil {
		glog.Errorf(ctx, "读取流数据失败: %v", err)
		return
	}

	g.Log().Debugf(ctx, "收到来自 %s 的消息: %v> ", peerID, string(buf[:n]))
}

// 发送数据到目标节点
func (s *sP2P) SendData(targetID string, data []byte) error {
	peerID, exists := s.client.peers[targetID]
	if !exists {
		return fmt.Errorf("未找到目标节点 %s 的连接", targetID)
	}

	// 创建流
	stream, err := s.client.host.NewStream(gctx.New(), peerID, protocol.ID("/p2p-chat/1.0.0"))
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
func (s *sP2P) receiveGatewayMessages() {
	for {
		_, data, err := s.client.wsConn.ReadMessage()
		if err != nil {
			glog.Errorf(ctx, "接收网关消息失败: %v", err)
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
			glog.Infof(ctx, "注册成功")

		case MsgTypeDiscoverAck:
			var msgData struct {
				Found  bool     `json:"found"`
				PeerID string   `json:"peer_id,omitempty"`
				Addrs  []string `json:"addrs,omitempty"`
			}
			if err = gjson.DecodeTo(msg.Data, &msgData); err != nil {
				g.Log().Errorf(ctx, "解析发现响应失败: %v", err)
				continue
			}

			if !msgData.Found {
				fmt.Println("未找到目标节点")
				continue
			}

			// 解析并连接目标节点
			peerID, err := peer.Decode(msgData.PeerID)
			if err != nil {
				glog.Errorf(ctx, "解析PeerID失败: %v", err)
				continue
			}

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
			}(peerID, msg.To)

		case MsgTypePunchRequest:
			err = s.handlePunchRequest(msg.Data)

		case MsgTypeError:
			var data struct {
				Error string `json:"error"`
			}
			json.Unmarshal(msg.Data, &data)
			glog.Errorf(ctx, "网关错误: %s", data.Error)
		}
	}
}
