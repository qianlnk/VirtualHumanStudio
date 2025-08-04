package storages

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"time"
)

const datetimeLayout = "2006-01-02 15:04:05"

var Client Storage
var UploadClient Storage // 文件直接上传云盘

var (
	//ErrNotInit 未初始化错误
	ErrNotInit = errors.New("storage does not init")
	//ErrNotExist 文件不存在
	ErrNotExist = errors.New("no such file or directory")
)

type (
	PutPolicy struct {
		UserID       string `yaml:"UserID"`
		Key          string `yaml:"Key"`
		InsertOnly   uint16 `yaml:"InsertOnly"`
		ReturnBody   string `yaml:"ReturnBody"`
		Expires      uint32 `yaml:"Expires"`
		SizeMax      int64  `yaml:"SizeMax"`
		CallbackURL  string `yaml:"CallbackURL"`
		CallbackHost string `yaml:"CallbackHost"`
	}

	CallbackConfig struct {
		URL      string `yaml:"URL"      json:"callbackUrl"`
		Host     string `yaml:"Host"     json:"callbackHost"`
		Body     string `yaml:"Body"     json:"callbackBody"`
		BodyType string `yaml:"BodyType" json:"callbackBodyType"`
	}

	FileInfo struct {
		// bucket名称
		Bucket string `json:"bucket"`
		// 文件名称
		Name string `json:"file_name"`
		// 文件扩展
		Ext string `json:"file_ext"`
		// 文件大小
		Size int64 `json:"file_size"`
		// 文件hash
		Hash string `json:"file_hash"`
		// 文件MIME类型
		MimeType string `json:"file_mimetype"`
		// 文件用户
		EndUser string `json:"end_user"`
		// 文件URL
		URL []string `json:"url"`
		// 创建时间
		CreatedAt int64 `json:"created_at"`
		// 更新时间
		UpdatedAt int64 `json:"updated_at"`
		// 审核状态
		AuditStatus string `json:"audit_status"`
		// 文件路径
		Path string `json:"path"`
	}

	Storage interface {
		// 初始化
		Init() (err error)
		//UploadToken 获取上传凭证
		UploadToken(ctx context.Context, policy *PutPolicy) (method string, host string, header map[string]string, form map[string]string, usage string, err error)
		// 保存文件
		SaveFile(ctx context.Context, key string, reader io.Reader) (err error)
		// 打开文件
		OpenFile(ctx context.Context, key string) (reader io.ReadCloser, err error)
		// 获取文件
		FetchFile(ctx context.Context, key string) (data []byte, err error)
		// 文件状态
		StatFile(ctx context.Context, key string) (info FileInfo, err error)
		// 删除文件
		Delete(ctx context.Context, key string) (err error)
		// 复制文件
		Copy(ctx context.Context, srcKey string, destKey string) (err error)
		// 移动文件
		Move(ctx context.Context, srcKey string, destKey string) (err error)
		// 遍历文件夹及其子目录中的所有文件
		List(ctx context.Context, prefix string, offset int, limit int) (infos []FileInfo, err error)
		// 创建软链接
		Symlink(ctx context.Context, srcKey string, destKey string) (err error)
		// 获取文件url
		GetFileUrl(ctx context.Context, key string) (url string, err error)
	}
)

func MapToFileInfo(info map[string]string, domain []string) FileInfo {
	size, _ := strconv.ParseInt(info["file_size"], 10, 64)
	createdAt, _ := time.ParseInLocation(datetimeLayout, info["created_at"], time.Local)
	updatedAt, _ := time.ParseInLocation(datetimeLayout, info["updated_at"], time.Local)

	fileInfo := FileInfo{
		Bucket:    info["bucket_name"],
		Name:      info["file_name"],
		Ext:       info["file_ext"],
		MimeType:  info["file_mimetype"],
		Size:      size,
		Hash:      info["file_hash"],
		EndUser:   info["user"],
		CreatedAt: createdAt.Unix(),
		UpdatedAt: updatedAt.Unix(),
	}

	fileInfo.MakeFileUrl(domain)
	return fileInfo
}

func (fileInfo *FileInfo) MakeFileUrl(domain []string) {
	urls := make([]string, 0)
	for _, item := range domain {
		url := fmt.Sprintf("%s/%s/%s/%s", item, url.QueryEscape(fileInfo.EndUser), url.QueryEscape(fileInfo.Bucket), url.QueryEscape(fileInfo.Name))
		urls = append(urls, url)
	}
	fileInfo.URL = urls
}
