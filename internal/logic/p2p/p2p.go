package p2p

import (
	"sync"

	"github.com/ayflying/p2p/internal/service"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/libp2p/go-libp2p/core/protocol"
)

var (
	ctx = gctx.New()
)

// 常量定义
const (
	ProtocolID  protocol.ID = "/ay/1.0.0"
	DefaultPort             = 51888
)

type MsgType string

const (
	MsgTypeRegister      = "register"       // 客户端注册
	MsgTypeRegisterAck   = "register_ack"   // 注册响应
	MsgTypeDiscover      = "discover"       // 发现目标客户端
	MsgTypeDiscoverAck   = "discover_ack"   // 发现响应
	MsgTypeConnectionReq = "connection_req" // 连接请求通知
	MsgTypePunchRequest  = "punch_request"
	MsgTypeError         = "error" // 错误消息
)

// 注册请求数据
type RegisterData struct {
	ClientID string `json:"client_id"`
}

// sP2P 结构体实现 IP2P 接口
type sP2P struct {
	Clients map[string]*ClientConn // 客户端ID -> 连接
	lock    sync.RWMutex

	client *Client
}

// New 创建一个新的 P2P 服务实例
func New() *sP2P {
	return &sP2P{
		Clients: make(map[string]*ClientConn),
		client:  &Client{},
	}
}

func init() {
	service.RegisterP2P(New())
}
