package p2p

import (
	"context"

	"github.com/ayflying/p2p/api/p2p/v1"
	"github.com/ayflying/p2p/internal/service"
)

func (c *ControllerV1) Connect(ctx context.Context, req *v1.ConnectReq) (res *v1.ConnectRes, err error) {
	err = service.P2P().DiscoverAndConnect(req.TargetID)
	return
}
