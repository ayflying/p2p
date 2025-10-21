package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gorilla/websocket"
)

var (
	Ws *websocket.Conn
)

// 客户端连接信息
type ClientConn struct {
	ID         string
	PeerID     string
	Addrs      []string
	Conn       *websocket.Conn
	LastActive time.Time
}

// 消息结构
type GatewayMessage struct {
	Type MsgType         `json:"type"`
	From string          `json:"from"`
	To   string          `json:"to,omitempty"`
	Data json.RawMessage `json:"data,omitempty"`
}

func (s *sP2P) GatewayStart(ctx context.Context, group *ghttp.RouterGroup) (err error) {
	var wsUpGrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		//// CheckOrigin allows any origin in development
		//// In production, implement proper origin checking for security
		//CheckOrigin: func(r *http.Request) bool {
		//	return true
		//},
		//// Error handler for upgrade failures
		//Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		//	// Implement error handling logic here
		//},
	}

	group.Bind(func(r *ghttp.Request) {
		// Upgrade HTTP connection to WebSocket
		ws, err2 := wsUpGrader.Upgrade(r.Response.Writer, r.Request, nil)
		if err2 != nil {
			r.Response.Write(err2.Error())
			return
		}
		defer ws.Close()

		// Message handling loop
		for {
			_, data, _err := ws.ReadMessage()
			if _err != nil {
				//g.Log().Errorf(ctx, "读取消息失败: %v", err)
				//s.sendError(ws, err.Error())
				break
			}

			var msg GatewayMessage
			if err = json.Unmarshal(data, &msg); err != nil {
				//g.Log().Error(ctx, "消息格式错误")
				s.sendError(ws, "消息格式错误")
				continue
			}

			// 处理不同类型的消息
			switch msg.Type {
			case MsgTypeRegister:
				s.handleRegister(ctx, ws, msg)
			case MsgTypeDiscover:
				s.handleDiscover(ctx, ws, msg)
			default:
				g.Log().Error(ctx, "未知消息类型: %s", msg.Type)
			}
		}
		// Log connection closure
		g.Log().Infof(ctx, "websocket %v connection closed", ws.RemoteAddr())

	})

	return
}

// 处理注册请求
func (s *sP2P) handleRegister(ctx context.Context, conn *websocket.Conn, msg GatewayMessage) {
	if msg.From == "" {
		g.Log().Error(ctx, "客户端ID不能为空")
		return
	}

	var data struct {
		PeerID string   `json:"peer_id"`
		Addrs  []string `json:"addrs"`
	}

	if err := json.Unmarshal(msg.Data, &data); err != nil {
		s.sendError(conn, "注册数据格式错误")
		return
	}

	// 追加公网ip
	publicIp, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	ParseIP := net.ParseIP(publicIp)
	var ipType string
	if ParseIP.To4() != nil {
		ipType = "ip4"
	} else {
		ipType = "ip6"
	}
	port2 := 53533
	data.Addrs = append(data.Addrs, fmt.Sprintf("/%s/%s/tcp/%d", ipType, publicIp, port2))
	data.Addrs = append(data.Addrs, fmt.Sprintf("/%s/%s/udp/%d/quic-v1", ipType, publicIp, port2))

	// 过滤回环地址
	data.Addrs = s.filterLoopbackAddrs(data.Addrs)

	// 保存客户端信息
	client := &ClientConn{
		ID:         msg.From,
		PeerID:     data.PeerID,
		Addrs:      data.Addrs,
		Conn:       conn,
		LastActive: time.Now(),
	}

	s.lock.Lock()
	s.Clients[msg.From] = client
	s.lock.Unlock()

	glog.Infof(ctx, "客户端 ip=%s,%s 注册成功，PeerID: %s", conn.RemoteAddr(), msg.From, data.PeerID)

	// 发送注册成功响应
	err := s.sendMessage(conn, GatewayMessage{
		Type: MsgTypeRegisterAck,
		Data: json.RawMessage(`{"success": true, "message": "注册成功"}`),
	})
	if err != nil {
		s.sendError(conn, err.Error())
	}
}

// 清理超时客户端
func (s *sP2P) cleanupClients(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			s.lock.Lock()
			for id, client := range s.Clients {
				if now.Sub(client.LastActive) > 60*time.Second {
					client.Conn.Close()
					delete(s.Clients, id)
					glog.Infof(ctx, "清理超时客户端: %s", id)
				}
			}
			s.lock.Unlock()
		}
	}
}

// 发送错误消息
func (s *sP2P) sendError(conn *websocket.Conn, errMsg string) {
	s.sendMessage(conn, GatewayMessage{
		Type: "error",
		Data: json.RawMessage(fmt.Sprintf(`{"error": "%s"}`, errMsg)),
	})
}

// 发送消息
func (s *sP2P) sendMessage(conn *websocket.Conn, msg GatewayMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		glog.Errorf(gctx.New(), "序列化消息失败: %v", err)
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}

// 处理发现请求
func (s *sP2P) handleDiscover(ctx context.Context, conn *websocket.Conn, msg GatewayMessage) {
	if msg.From == "" {
		s.sendError(conn, "消息缺少发送方ID（from）")
		return
	}

	var data struct {
		TargetID string `json:"target_id"`
	}

	if err := json.Unmarshal(msg.Data, &data); err != nil {
		s.sendError(conn, "发现请求格式错误，需包含target_id")
		return
	}

	if data.TargetID == "" {
		s.sendError(conn, "目标ID不能为空")
		return
	}

	// 获取请求方和目标方信息
	s.lock.RLock()
	//fromClient, fromExists := s.Clients[msg.From]
	//targetClient, targetExists := s.Clients[data.TargetID]
	fromClient := s.Clients[msg.From]        // 发送方：msg.From
	targetClient := s.Clients[data.TargetID] // 目标方：data.TargetID
	s.lock.RUnlock()

	//if !fromExists {
	//	s.sendError(conn, "请先注册")
	//	return
	//}

	// 更新活动时间
	s.lock.Lock()
	fromClient.LastActive = time.Now()
	s.lock.Unlock()

	if targetClient == nil {
		// 目标不存在
		s.sendMessage(conn, GatewayMessage{
			Type: MsgTypeDiscoverAck,
			From: "gateway",
			To:   msg.From,
			//Data: json.RawMessage(`{"found": false}`),
			Data: gjson.MustEncode(g.Map{
				"found": false,
			}),
		})
		return
	}

	// 向请求方发送目标信息
	s.sendMessage(conn, GatewayMessage{
		Type: MsgTypeDiscoverAck,
		From: "gateway", // 发送方是网关
		To:   msg.From,  // 接收方是原请求方
		Data: gjson.MustEncode(g.Map{
			"found":     true,
			"peer_id":   targetClient.PeerID,
			"addrs":     targetClient.Addrs,
			"target_id": data.TargetID,
		}),
	})

	// 向目标方发送打洞请求（协调时机）
	s.sendMessage(targetClient.Conn, GatewayMessage{
		Type: MsgTypePunchRequest,
		From: msg.From,      // 发送方是原请求方
		To:   data.TargetID, // 接收方是目标方
		Data: gjson.MustEncode(g.Map{
			"from_id": msg.From,
			"peer_id": fromClient.PeerID,
			//"addrs":   s.getAddrsJSON(fromClient.Addrs),
			"addrs": fromClient.Addrs,
		}),
	})
}

// 获取地址JSON字符串
func (s *sP2P) getAddrsJSON(addrs []string) string {
	strs := make([]string, len(addrs))
	for i, addr := range addrs {
		strs[i] = fmt.Sprintf("%q", addr)
	}
	return fmt.Sprintf("[%s]", strings.Join(strs, ","))
}
