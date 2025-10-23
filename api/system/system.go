// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package system

import (
	"context"

	"github.com/ayflying/p2p/api/system/v1"
)

type ISystemV1 interface {
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error)
}
