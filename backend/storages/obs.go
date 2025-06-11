package storages

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
)

type OBSConfig struct {
	Bucket    string `yaml:"Bucket"    json:"bucket"`
	AccessKey string `yaml:"AccessKey" json:"access_key"`
	SecretKey string `yaml:"SecretKey" json:"secret_key"`
	Endpoint  string `yaml:"Endpoint"  json:"endpoint"`
	BasePath  string `yaml:"BasePath"  json:"base_path"`
}

type OBS struct {
	config OBSConfig
	client *obs.ObsClient
}

func NewOBS(cfg OBSConfig) (*OBS, error) {
	client, err := obs.New(cfg.AccessKey, cfg.SecretKey, cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	return &OBS{config: cfg, client: client}, nil
}

func (o *OBS) Init() error {
	if o.client == nil {
		return ErrNotInit
	}

	return nil
}

func (o *OBS) UploadToken(ctx context.Context, policy *PutPolicy) (method string, host string, header map[string]string, form map[string]string, usage string, err error) {
	input := &obs.CreateSignedUrlInput{
		Method:  http.MethodPut,
		Bucket:  o.config.Bucket,
		Key:     path.Join(o.config.BasePath, policy.Key),
		Expires: int(policy.Expires),
	}

	uri, err := o.client.CreateSignedUrl(input)
	if err != nil {
		return "", "", nil, nil, "", err
	}

	method = http.MethodPut
	usage = fmt.Sprintf(`curl -X '%s' '%s'`, method, uri.SignedUrl)

	return method, uri.SignedUrl, nil, nil, usage, nil
}

func (o *OBS) SaveFile(ctx context.Context, key string, reader io.Reader) (err error) {
	key = path.Join(o.config.BasePath, key)

	buffer := make([]byte, 512)
	n, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		return err
	}

	realContentType := http.DetectContentType(buffer[:n])
	realContentType = strings.Split(realContentType, ";")[0]

	// 重置reader，将已读取的部分与剩余内容合并
	fullReader := io.MultiReader(bytes.NewReader(buffer[:n]), reader)

	fmt.Println("===", o.config.Bucket, key, realContentType)

	input := &obs.PutObjectInput{}

	input.Bucket = o.config.Bucket
	input.Key = key
	input.Body = fullReader
	input.ContentType = realContentType

	_, err = o.client.PutObject(input)
	if err != nil {
		return err
	}

	return nil
}

func (o *OBS) FetchFile(ctx context.Context, key string) (data []byte, err error) {
	key = path.Join(o.config.BasePath, key)

	input := &obs.GetObjectInput{}

	input.Bucket = o.config.Bucket
	input.Key = key

	output, err := o.client.GetObject(input)
	if err != nil {
		return nil, err
	}

	defer output.Body.Close()

	data, err = io.ReadAll(output.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (o *OBS) OpenFile(ctx context.Context, key string) (reader io.ReadCloser, err error) {
	key = path.Join(o.config.BasePath, key)

	input := &obs.GetObjectInput{}

	input.Bucket = o.config.Bucket
	input.Key = key

	output, err := o.client.GetObject(input)
	if err != nil {
		return nil, err
	}

	return output.Body, nil
}

func (o *OBS) StatFile(ctx context.Context, key string) (info FileInfo, err error) {
	key = path.Join(o.config.BasePath, key)

	input := &obs.GetAttributeInput{}
	input.Bucket = o.config.Bucket
	input.Key = key

	output, err := o.client.GetAttribute(input)
	if err != nil {
		return info, err
	}

	info.Size = output.ContentLength
	info.Hash = output.ETag
	info.MimeType = output.ContentType

	return info, nil
}

func (o *OBS) Delete(ctx context.Context, key string) (err error) {
	key = path.Join(o.config.BasePath, key)

	input := &obs.DeleteObjectInput{}

	input.Bucket = o.config.Bucket
	input.Key = key

	_, err = o.client.DeleteObject(input)
	if err != nil {
		return err
	}

	return nil
}

func (o *OBS) DeleteFileList(ctx context.Context, keys []string) (err error) {
	input := &obs.DeleteObjectsInput{}

	input.Bucket = o.config.Bucket
	input.Objects = make([]obs.ObjectToDelete, 0, len(keys))
	for _, key := range keys {
		input.Objects = append(input.Objects, obs.ObjectToDelete{
			Key: path.Join(o.config.BasePath, key),
		})
	}

	_, err = o.client.DeleteObjects(input)
	if err != nil {
		return err
	}

	return nil
}

func (o *OBS) Copy(ctx context.Context, srcKey string, destKey string) (err error) {
	input := &obs.CopyObjectInput{}

	input.Bucket = o.config.Bucket
	input.Key = path.Join(o.config.BasePath, destKey)
	input.CopySourceBucket = o.config.Bucket
	input.CopySourceKey = path.Join(o.config.BasePath, srcKey)

	_, err = o.client.CopyObject(input)
	if err != nil {
		return err
	}

	return nil
}

func (o *OBS) Move(ctx context.Context, srcKey string, destKey string) (err error) {
	err = o.Copy(ctx, srcKey, destKey)
	if err != nil {
		return err
	}

	err = o.Delete(ctx, srcKey)
	if err != nil {
		return err
	}

	return nil
}

func (o *OBS) List(ctx context.Context, prefix string, offset int, limit int) (infos []FileInfo, err error) {
	prefix = path.Join(o.config.BasePath, prefix)

	input := &obs.ListObjectsInput{}

	input.Bucket = o.config.Bucket
	input.Prefix = prefix
	input.MaxKeys = offset + limit

	output, err := o.client.ListObjects(input)
	if err != nil {
		return nil, err
	}

	for _, object := range output.Contents {
		infos = append(infos, FileInfo{
			Bucket: o.config.Bucket,
			Name:   object.Key,
			Size:   object.Size,
			Hash:   object.ETag,
		})
	}

	return infos, nil
}

func (o *OBS) GetFileUrl(ctx context.Context, key string) (url string, err error) {
	if key == "" {
		return "", nil
	}

	key = path.Join(o.config.BasePath, key)

	input := &obs.CreateSignedUrlInput{}

	input.Method = http.MethodGet
	input.Bucket = o.config.Bucket
	input.Key = key
	input.Expires = 300

	uri, err := o.client.CreateSignedUrl(input)
	if err != nil {
		return "", err
	}

	return uri.SignedUrl, nil
}

func (o *OBS) Symlink(ctx context.Context, srcKey string, destKey string) (err error) {
	err = o.Copy(ctx, srcKey, destKey)
	return
}
