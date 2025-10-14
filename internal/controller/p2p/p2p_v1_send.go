package p2p

import (
	"context"

	"github.com/ayflying/p2p/api/p2p/v1"
	"github.com/ayflying/p2p/internal/service"
)

func (c *ControllerV1) Send(ctx context.Context, req *v1.SendReq) (res *v1.SendRes, err error) {
	err = service.P2P().SendData(req.TargetID, []byte(req.Data))
	return
}
