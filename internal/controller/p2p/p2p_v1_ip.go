package p2p

import (
	"context"

	"github.com/ayflying/p2p/api/p2p/v1"
	"github.com/gogf/gf/v2/frame/g"
)

func (c *ControllerV1) Ip(ctx context.Context, req *v1.IpReq) (res *v1.IpRes, err error) {
	ip := g.RequestFromCtx(ctx).GetRemoteIp()
	g.RequestFromCtx(ctx).Response.Write(ip)
	return
}
