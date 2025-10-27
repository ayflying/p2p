package system

import (
	"context"
	"net/url"
	"path"

	"github.com/ayflying/p2p/api/system/v1"
	"github.com/ayflying/p2p/internal/service"
	"github.com/gogf/gf/v2/crypto/gsha1"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"
)

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	getRunFile := gcmd.GetArg(0).String()
	fileSha, err := gsha1.EncryptFile(getRunFile)
	g.Log().Debugf(ctx, "当前文件哈希值：%v", fileSha)

	versionUrl, _ := url.JoinPath(req.Url, "version.json")
	resp, err := g.Client().Get(ctx, versionUrl)
	var version map[string]string
	gjson.DecodeTo(resp.ReadAll(), &version)

	for k, _ := range version {
		//downloadUrl, _ := url.QueryUnescape(v)
		downloadUrl, _ := url.JoinPath(req.Url, req.Version, k+".gz")
		fileByte, err2 := g.Client().Get(ctx, downloadUrl)
		if err2 != nil {
			g.Log().Error(ctx, err2)
			continue
		}
		putFile := path.Join("download", gfile.Basename(downloadUrl))
		err2 = gfile.PutBytes(putFile, fileByte.ReadAll())
		if err2 != nil {
			g.Log().Error(ctx, err2)
			continue
		}
	}

	//更新文件
	err = service.System().Update(ctx)
	type DataType struct {
		File []byte `json:"file"`
		Name string `json:"name"`
	}

	var msgData = struct {
		Files []*DataType `json:"files"`
	}{}

	msgData.Files = []*DataType{}

	files, _ := gfile.ScanDir("download", ".*gz")

	for _, v := range files {
		msgData.Files = append(msgData.Files, &DataType{
			File: gfile.GetBytes(v),
			Name: gfile.Basename(v),
		})
	}

	service.P2P().SendAll("update", msgData)
	return
}
