package p2p

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

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

// 获取公网IP并判断类型（ipv4/ipv6）
func (s *sP2P) getPublicIPAndType() (ip string, ipType string, err error) {
	// 公网IP查询接口（多个备用）
	apis := []string{
		"https://api.ip.sb/ip",
		"https://ip.3322.net",
		"https://ifconfig.cn",
	}

	client := http.Client{Timeout: 5 * time.Second}
	for _, api := range apis {
		resp, err := client.Get(api)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		// 读取响应（纯IP字符串）
		buf := make([]byte, 128)
		n, err := resp.Body.Read(buf)
		if err != nil {
			continue
		}

		ip = strings.TrimSpace(string(buf[:n]))
		if ip == "" {
			continue
		}

		// 判断IP类型
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			continue // 无效IP格式
		}

		if parsedIP.To4() != nil {
			return ip, "ipv4", nil // IPv4
		} else if parsedIP.To16() != nil {
			return ip, "ipv6", nil // IPv6
		}
	}

	return "", "", fmt.Errorf("所有公网IP查询接口均失败")
}

// 只获取IPv4公网IP（过滤IPv6结果）
func (s *sP2P) getIPv4PublicIP() (string, error) {
	// 优先使用只返回IPv4的API，避免IPv6干扰
	ipv4OnlyAPIs := []string{
		//"https://api.ip.sb/ip",
		"https://ip.3322.net",
		//"https://ifconfig.cn",
	}

	client := http.Client{Timeout: 5 * time.Second}
	for _, api := range ipv4OnlyAPIs {
		resp, err := client.Get(api)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		// 读取响应
		buf := make([]byte, 128)
		n, err := resp.Body.Read(buf)
		if err != nil {
			continue
		}

		ip := strings.TrimSpace(string(buf[:n]))
		if ip == "" {
			continue
		}

		// 严格验证是否为IPv4（过滤IPv6）
		parsedIP := net.ParseIP(ip)
		if parsedIP != nil && parsedIP.To4() != nil { // 确保是IPv4
			return ip, nil
		}
	}

	return "", fmt.Errorf("所有IPv4公网查询接口均失败或返回非IPv4地址")
}

// 过滤地址列表，排除127.0.0.1回环地址
func (s *sP2P) filterLoopbackAddrs(addrStrs []string) []string {
	var filtered []string
	for _, addrStr := range addrStrs {
		// 直接过滤包含127.0.0.1的地址字符串
		if strings.Contains(addrStr, "/ip4/127.0.0.1/") {
			continue // 跳过回环地址
		}
		filtered = append(filtered, addrStr)
	}

	// 移除重复地址
	filtered = s.removeDuplicates(filtered)
	return filtered
}

// 去除字符串切片中的重复元素，保持首次出现的顺序
func (s *sP2P) removeDuplicates(strs []string) []string {
	seen := make(map[string]bool)          // 用于记录已出现的字符串
	result := make([]string, 0, len(strs)) // 结果切片，预分配容量

	for _, s := range strs {
		if !seen[s] { // 如果未出现过
			seen[s] = true             // 标记为已出现
			result = append(result, s) // 添加到结果
		}
	}
	return result
}
