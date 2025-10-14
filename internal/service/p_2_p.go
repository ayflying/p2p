// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"
)

type (
	IP2P interface {
		Start(ctx context.Context, wsStr string) (err error)
		// 发现并连接目标节点
		DiscoverAndConnect(targetID string) error
		// 发送数据到目标节点
		SendData(targetID string, data []byte) error
		GatewayStart(ctx context.Context, group *ghttp.RouterGroup) (err error)
	}
)

var (
	localP2P IP2P
)

func P2P() IP2P {
	if localP2P == nil {
		panic("implement not found for interface IP2P, forgot register?")
	}
	return localP2P
}

func RegisterP2P(i IP2P) {
	localP2P = i
}
