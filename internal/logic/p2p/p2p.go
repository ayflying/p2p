package p2p

import (
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ayflying/p2p/internal/service"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/util/grand"
	"github.com/libp2p/go-libp2p/core/crypto"
)

var (
	//ctx = gctx.New()
	ip string
)

// 常量定义
const (
	ProtocolID  string = "/ay"
	DefaultPort        = 51888
)

var (
	ipAPIs = []string{
		//"http://ay.cname.com:51888/p2p/ip",
		"http://54.67.8.27:51888/p2p/ip",
	}
)

type MsgType string

const (
	MsgTypeRegister      = "register"       // 客户端注册
	MsgTypeRegisterAck   = "register_ack"   // 注册响应
	MsgTypeDiscover      = "discover"       // 发现目标客户端
	MsgTypeDiscoverAck   = "discover_ack"   // 发现响应
	MsgTypeConnectionReq = "connection_req" // 连接请求通知
	MsgTypePunchRequest  = "punch_request"  // 打洞请求
	MsgTypeError         = "error"          // 错误消息
	MsgUpdate            = "update"         // 更新节点信息
)

// 注册请求数据
type RegisterData struct {
	ClientID string `json:"client_id"`
}

// sP2P 结构体实现 IP2P 接口
type sP2P struct {
	Clients map[string]*ClientConn // 客户端ID -> 连接
	lock    sync.RWMutex
	dht     *DHTType
	privKey crypto.PrivKey
	client  *Client
}

// New 创建一个新的 P2P 服务实例
func New() *sP2P {

	return &sP2P{
		Clients: make(map[string]*ClientConn),
		client:  &Client{},
		dht:     &DHTType{},
	}
}

func init() {
	service.RegisterP2P(New())
	ip, _ = service.P2P().GetIPv4PublicIP()
}

// 获取公网IP并判断类型（ipv4/ipv6）
func (s *sP2P) getPublicIPAndType() (ip string, ipType string, err error) {
	// 公网IP查询接口（多个备用）

	client := http.Client{Timeout: 5 * time.Second}
	for _, api := range ipAPIs {
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
func (s *sP2P) GetIPv4PublicIP() (string, error) {
	ctx := gctx.New()
	// 优先使用只返回IPv4的API，避免IPv6干扰

	//client := http.Client{Timeout: 5 * time.Second}
	for _, api := range ipAPIs {
		//resp, err := client.Get(api)
		resp, err := g.Client().Timeout(5*time.Second).Get(ctx, api)
		if err != nil {
			continue
		}
		defer resp.Close()
		//defer resp.Body.Close()

		// 读取响应
		//buf := make([]byte, 128)
		//n, err := resp.Body.Read(buf)
		//if err != nil {
		//	continue
		//}

		//ip := strings.TrimSpace(string(buf[:n]))
		ip := strings.TrimSpace(resp.ReadAllString())
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

// 生成固定密钥（核心：通过固定种子生成相同密钥）
func (s *sP2P) generateFixedKey() (crypto.PrivKey, error) {
	privKeyPath := "runtime/p2p.key"
	if ok := gfile.Exists(privKeyPath); ok {
		// 从文件读取密钥
		keyBytes := gfile.GetBytes(privKeyPath)
		// 2. 解析PEM块（关键：提取真正的私钥数据）
		pemBlock, _ := pem.Decode(keyBytes)
		if pemBlock == nil {
			return nil, fmt.Errorf("私钥文件格式错误（非PEM格式）")
		}
		privKey, err := crypto.UnmarshalPrivateKey(pemBlock.Bytes)
		if err != nil {
			return nil, err
		}
		return privKey, nil
	}

	// 固定种子（修改此种子可生成不同的固定密钥）
	var fixedSeed = []byte(grand.S(10)) // 自定义固定种子

	// 用固定种子初始化随机数生成器
	seed := binary.BigEndian.Uint64(fixedSeed[:8]) // 取种子前8字节作为随机数种子
	r := rand.New(rand.NewSource(int64(seed)))

	// 生成ED25519密钥对（基于固定种子，每次生成结果相同）
	privKey, _, err := crypto.GenerateEd25519Key(r)
	keyBytes, err := crypto.MarshalPrivateKey(privKey)
	// 用PEM格式包装（标准格式，便于存储和解析）
	pemBlock := &pem.Block{
		Type:  "LIBP2P PRIVATE KEY", // 标识为libp2p私钥
		Bytes: keyBytes,
	}
	err = gfile.PutBytes(privKeyPath, pem.EncodeToMemory(pemBlock))
	if err != nil {
		panic(fmt.Sprintf("保存私钥失败: %v", err))
	}
	//if err := os.WriteFile(privKeyPath, pem.EncodeToMemory(pemBlock), 0600); err != nil {
	//	panic(fmt.Sprintf("保存私钥失败: %v", err))
	//}

	fmt.Println("私钥生成成功，文件路径：", privKeyPath)

	return privKey, err
}
