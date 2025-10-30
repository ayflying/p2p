package cmd

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"runtime"

	systemV1 "github.com/ayflying/p2p/api/system/v1"
	"github.com/ayflying/p2p/internal/service"
	"github.com/gogf/gf/v2/crypto/gsha1"
	"github.com/gogf/gf/v2/encoding/gcompress"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"
)

var (
	Update = gcmd.Command{
		Name:  "update",
		Usage: "update",
		Brief: "更新版本",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			g.Log().Info(ctx, "准备上传更新文件")
			//加载编辑配置文件
			g.Cfg("hack").GetAdapter().(*gcfg.AdapterFile).SetFileName("hack/config.yaml")
			//获取文件名
			getName, err := g.Cfg("hack").Get(ctx, "gfcli.build.name")
			name := getName.String()

			getPath, err := g.Cfg("hack").Get(ctx, "gfcli.build.path")
			pathMain := getPath.String()

			//获取版本号
			getVersion, err := g.Cfg("hack").Get(ctx, "gfcli.build.version")
			version := getVersion.String()

			// 拼接操作系统和架构（格式：OS_ARCH）
			platform := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)

			rootDir := "server_update"

			var versionFile = make(map[string]string)
			var filePath = path.Join(pathMain, version, platform, name)
			dirList, _ := gfile.ScanDir(path.Join(pathMain, version), "*", false)
			for _, v := range dirList {
				updatePlatform := gfile.Name(v)
				updateFilePath := path.Join(rootDir, name, version, updatePlatform)

				var obj bytes.Buffer
				g.Log().Debugf(ctx, "读取目录成功:%v", v)
				fileMian := path.Join(v, name)
				g.Log().Debugf(ctx, "判断当前文件是否存在：%v", fileMian)
				if gfile.IsFile(fileMian) {
					// 写入文件哈希
					versionFile[updatePlatform] = gsha1.MustEncryptFile(fileMian)
					err = gcompress.GzipPathWriter(fileMian, &obj)
					service.S3().PutObject(ctx, &obj, updateFilePath+".gz")
					g.Log().Debugf(ctx, "成功上传文件到：%v", updateFilePath+".gz")
				}
				if gfile.IsFile(fileMian + ".exe") {
					// 写入文件哈希
					versionFile[updatePlatform] = gsha1.MustEncryptFile(fileMian + ".exe")
					err = gcompress.GzipPathWriter(fileMian+".exe", &obj)
					service.S3().PutObject(ctx, &obj, updateFilePath+".gz")
					g.Log().Debugf(ctx, "成功上传文件到：%v", updateFilePath+".gz")
				}

				// 写入文件版本文件
				fileByte := gjson.MustEncode(versionFile)
				service.S3().PutObject(ctx, bytes.NewReader(fileByte), path.Join(rootDir, name, "version.json"))
				if err != nil {
					g.Log().Error(ctx, err)
				}
			}
			g.Log().Debugf(ctx, "当前获取到的地址为：%v", filePath)

			versionUrl := service.S3().GetCdnUrl(path.Join(rootDir, name))
			listVar := g.Cfg().MustGet(ctx, "message.list")
			var p2pItem []struct {
				Host string `json:"host"`
				Port int    `json:"port"`
				SSL  bool   `json:"ssl"`
				Ws   string `json:"ws"`
			}
			listVar.Scan(&p2pItem)
			for _, v := range p2pItem {

				url := "http"
				if v.SSL == true {
					url = "https"
				}
				url = fmt.Sprintf("%s://%s:%d/system/update", url, v.Host, v.Port)

				g.Log().Debugf(ctx, "开始上传到服务器：%v,file=%v", url, versionUrl)
				_, err := g.Client().Get(ctx, url, systemV1.UpdateReq{
					Url:     versionUrl,
					Version: version,
				})
				if err != nil {
					g.Log().Error(ctx, err)
				}
			}
			return
		}}
)
