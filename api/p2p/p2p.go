// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package p2p

import (
	"context"

	"github.com/ayflying/p2p/api/p2p/v1"
)

type IP2PV1 interface {
	Connect(ctx context.Context, req *v1.ConnectReq) (res *v1.ConnectRes, err error)
	Send(ctx context.Context, req *v1.SendReq) (res *v1.SendRes, err error)
}
