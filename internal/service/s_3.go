// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
)

type (
	IS3 interface {
		// GetFileUrl 生成指向 S3 存储桶中指定文件的预签名 URL
		// 预签名 URL 可用于在有限时间内访问 S3 存储桶中的文件
		// 支持从缓存中获取预签名 URL，以减少重复请求
		GetFileUrl(ctx context.Context, name string, _expires ...time.Duration) (presignedURL *url.URL, err error)
		// PutFileUrl 生成一个用于上传文件到指定存储桶的预签名 URL
		// 预签名 URL 的有效期默认为 10 分钟
		PutFileUrl(ctx context.Context, name string) (presignedURL *url.URL, err error)
		// ListBuckets 获取当前 S3 客户端可访问的所有存储桶列表
		// 出错时返回 nil
		ListBuckets(ctx context.Context) []minio.BucketInfo
		// PutObject 上传文件到指定的存储桶中
		// 支持指定文件大小，未指定时将读取文件直到结束
		PutObject(ctx context.Context, f io.Reader, name string, _size ...int64) (res minio.UploadInfo, err error)
		// RemoveObject 从指定存储桶中删除指定名称的文件
		RemoveObject(ctx context.Context, name string) (err error)
		// ListObjects 获取指定存储桶中指定前缀的文件列表
		// 返回一个包含文件信息的通道
		ListObjects(ctx context.Context, prefix string) (res <-chan minio.ObjectInfo, err error)
		// StatObject 获取指定存储桶中指定文件的元数据信息
		StatObject(ctx context.Context, objectName string) (res minio.ObjectInfo, err error)
		// GetUrl 获取文件的访问地址
		// 支持返回默认文件地址，根据 SSL 配置生成不同格式的 URL
		GetUrl(filePath string, defaultFile ...string) (url string)
		// GetPath 从文件访问 URL 中提取文件路径
		GetPath(url string) (filePath string)
		// GetCdnUrl 通过文件名，获取直连地址
		GetCdnUrl(file string) string
		// CopyObject 在指定存储桶内复制文件
		// bucketName 存储桶名称
		// dstStr 目标文件路径
		// srcStr 源文件路径
		// 返回操作过程中可能出现的错误
		CopyObject(ctx context.Context, dstStr string, srcStr string) (err error)
		// Rename 重命名文件
		Rename(ctx context.Context, oldName string, newName string) (err error)
	}
)

var (
	localS3 IS3
)

func S3() IS3 {
	if localS3 == nil {
		panic("implement not found for interface IS3, forgot register?")
	}
	return localS3
}

func RegisterS3(i IS3) {
	localS3 = i
}
