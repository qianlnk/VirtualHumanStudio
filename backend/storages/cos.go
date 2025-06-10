package storages

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// COSConfig tencent cos config
type COSConfig struct {
	Bucket    string `yaml:"Bucket"    json:"bucket"`
	Appid     string `yaml:"Appid"     json:"appid"`
	Region    string `yaml:"Region"    json:"region"`
	SecretID  string `yaml:"SecretID"  json:"secret_id"`
	SecretKey string `yaml:"SecretKey" json:"secret_key"`
}

// COS tencent cos client
type COS struct {
	config COSConfig
	client *cos.Client
}

// NewCOS create cos client
func NewCOS(cfg COSConfig) (*COS, error) {
	rawurl := fmt.Sprintf("https://%s-%s.cos.%s.myqcloud.com", cfg.Bucket, cfg.Appid, cfg.Region)
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	b := &cos.BaseURL{BucketURL: u}

	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.SecretID,
			SecretKey: cfg.SecretKey,
		},
	})

	return &COS{
		config: cfg,
		client: client,
	}, nil
}

// Init isinit
func (c *COS) Init() error {
	if c.client == nil {
		return ErrNotInit
	}

	return nil
}

// UploadToken 获取上传凭证
func (c *COS) UploadToken(ctx context.Context, policy *PutPolicy) (method string, host string, header map[string]string, form map[string]string, usage string, err error) {
	method = http.MethodPut
	uri, err := c.client.Object.GetPresignedURL(ctx, method, policy.Key, c.config.SecretID, c.config.SecretKey, time.Second*time.Duration(policy.Expires), nil)
	if err != nil {
		return "", "", nil, nil, "", err
	}

	usage = fmt.Sprintf(`curl -X '%s' '%s'`, method, uri.String())

	return method, uri.String(), nil, nil, usage, nil
}

// SaveFile 保存文件
func (c *COS) SaveFile(ctx context.Context, key string, reader io.Reader) (err error) {
	// 检测内容类型（只读取文件开头部分）
	buffer := make([]byte, 512)
	n, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		return err
	}

	realContentType := http.DetectContentType(buffer[:n])
	realContentType = strings.Split(realContentType, ";")[0]

	// 重置reader，将已读取的部分与剩余内容合并
	fullReader := io.MultiReader(bytes.NewReader(buffer[:n]), reader)

	// 获取预签名URL
	method := http.MethodPut
	uri, err := c.client.Object.GetPresignedURL(ctx, method, key, c.config.SecretID, c.config.SecretKey, time.Second*300, nil)
	if err != nil {
		return err
	}

	// 创建请求，使用流式reader而不是加载整个文件到内存
	req, err := http.NewRequest(method, uri.String(), fullReader)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", realContentType)

	// 使用默认客户端发送请求
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	// 只读取响应体，不需要存储整个响应数据
	if res.StatusCode != 200 {
		respData, _ := io.ReadAll(res.Body)
		err = fmt.Errorf("%d %s", res.StatusCode, string(respData))
		return err
	}

	// 丢弃响应体
	_, err = io.Copy(io.Discard, res.Body)
	return err
}

// FetchFile 获取文件
func (c *COS) FetchFile(ctx context.Context, key string) (data []byte, err error) {
	res, err := c.client.Object.Get(ctx, key, nil)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	data, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return
}

// OpenFile 打开文件
func (c *COS) OpenFile(ctx context.Context, key string) (reader io.ReadCloser, err error) {
	res, err := c.client.Object.Get(ctx, key, nil)
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

// StatFile 文件状态
func (c *COS) StatFile(ctx context.Context, key string) (info FileInfo, err error) {
	res, err := c.client.Object.Head(ctx, key, nil)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			err = ErrNotExist
		}
		return info, err
	}
	defer res.Body.Close()

	info.Size, _ = strconv.ParseInt(res.Header.Get("Content-Length"), 10, 64)
	info.Hash = strings.Trim(res.Header.Get("Etag"), "\"")
	info.MimeType = res.Header.Get("Content-Type")
	return
}

// DeleteFile 删除文件
func (c *COS) Delete(ctx context.Context, key string) (err error) {
	_, err = c.client.Object.Delete(ctx, key)

	return
}

func (c *COS) DeleteFileList(ctx context.Context, keys []string) error {
	objs := make([]cos.Object, 0, len(keys))

	for _, k := range keys {
		objs = append(objs, cos.Object{
			Key: k,
		})
	}

	_, _, err := c.client.Object.DeleteMulti(ctx, &cos.ObjectDeleteMultiOptions{
		Quiet:   true,
		Objects: objs,
	})

	if err != nil {
		return err
	}

	return nil
}

// Copy 复制文件
func (c *COS) Copy(ctx context.Context, srcKey string, destKey string) (err error) {
	source := fmt.Sprintf("%s/%s", c.client.BaseURL.BucketURL.Host, srcKey)

	_, _, err = c.client.Object.Copy(ctx, destKey, source, nil)
	return
}

// Move 移动文件
func (c *COS) Move(ctx context.Context, srcKey string, destKey string) (err error) {
	err = c.Copy(ctx, srcKey, destKey)
	if err != nil {
		return
	}

	err = c.Delete(ctx, srcKey)
	return
}

// List 列出文件
func (c *COS) List(ctx context.Context, prefix string, offset int, limit int) (infos []FileInfo, err error) {
	// TODO
	opt := &cos.BucketGetOptions{
		Prefix:  prefix,
		MaxKeys: offset + limit,
	}

	c.client.Bucket.Get(ctx, opt)
	return
}

// GetFileUrl 获取文件url
func (c *COS) GetFileUrl(ctx context.Context, key string) (url string, err error) {
	u, err := c.client.Object.GetPresignedURL(ctx, http.MethodGet, key, c.config.SecretID, c.config.SecretKey, time.Second*300, nil)
	if err != nil {
		return "", err
	}

	return u.String(), nil
}

func (c *COS) Symlink(ctx context.Context, srcKey string, destKey string) (err error) {
	_, _, err = c.client.Object.Copy(ctx, destKey, srcKey, nil)
	return
}
