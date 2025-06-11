package storages

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

type TOSConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
}

type TOS struct {
	config TOSConfig
	client *tos.ClientV2
}

func NewTOS(config TOSConfig) (*TOS, error) {
	client, err := tos.NewClientV2(config.Endpoint, tos.WithRegion(config.Region), tos.WithCredentials(tos.NewStaticCredentials(config.AccessKey, config.SecretKey)))
	if err != nil {
		return nil, err
	}
	return &TOS{
		config: config,
		client: client,
	}, nil
}

func (t *TOS) Init() error {
	if t.client == nil {
		return ErrNotInit
	}
	return nil
}

func (t *TOS) UploadToken(ctx context.Context, policy *PutPolicy) (method string, host string, header map[string]string, form map[string]string, usage string, err error) {
	method = http.MethodPut
	input := &tos.PreSignedURLInput{
		Bucket:     t.config.Bucket,
		HTTPMethod: http.MethodPut,
		Key:        policy.Key,
		Expires:    int64(policy.Expires),
	}
	uri, err := t.client.PreSignedURL(input)
	if err != nil {
		return "", "", nil, nil, "", err
	}
	usage = fmt.Sprintf(`curl -X '%s' '%s'`, method, uri.SignedUrl)
	return method, uri.SignedUrl, nil, nil, usage, nil
}

func (t *TOS) SaveFile(ctx context.Context, key string, reader io.Reader) (err error) {
	buffer := make([]byte, 512)
	n, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		return err
	}

	realContentType := http.DetectContentType(buffer[:n])
	realContentType = strings.Split(realContentType, ";")[0]

	// 重置reader，将已读取的部分与剩余内容合并
	fullReader := io.MultiReader(bytes.NewReader(buffer[:n]), reader)

	fmt.Println("===", t.config.Bucket, key, realContentType)

	input := &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:      t.config.Bucket,
			Key:         key,
			ContentType: realContentType,
		},
		Content: fullReader,
	}
	output, err := t.client.PutObjectV2(ctx, input)
	if err != nil {
		return err
	}

	if output.StatusCode != http.StatusOK {
		return fmt.Errorf("save file failed, status code: %d", output.StatusCode)
	}

	return nil
}

func (t *TOS) FetchFile(ctx context.Context, key string) (data []byte, err error) {
	input := &tos.GetObjectV2Input{
		Bucket: t.config.Bucket,
		Key:    key,
	}
	output, err := t.client.GetObjectV2(ctx, input)
	if err != nil {
		return nil, err
	}

	defer output.Content.Close()

	data, err = io.ReadAll(output.Content)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (t *TOS) OpenFile(ctx context.Context, key string) (reader io.ReadCloser, err error) {
	input := &tos.GetObjectV2Input{
		Bucket: t.config.Bucket,
		Key:    key,
	}
	output, err := t.client.GetObjectV2(ctx, input)
	if err != nil {
		return nil, err
	}

	return output.Content, nil
}

func (t *TOS) StatFile(ctx context.Context, key string) (info FileInfo, err error) {
	input := &tos.HeadObjectV2Input{
		Bucket: t.config.Bucket,
		Key:    key,
	}
	output, err := t.client.HeadObjectV2(ctx, input)
	if err != nil {
		return info, err
	}

	info.Size = output.ContentLength
	info.Hash = output.ETag
	info.MimeType = output.ContentType
	return info, nil
}

func (t *TOS) Delete(ctx context.Context, key string) (err error) {
	input := &tos.DeleteObjectV2Input{
		Bucket: t.config.Bucket,
		Key:    key,
	}
	output, err := t.client.DeleteObjectV2(ctx, input)
	if err != nil {
		return err
	}

	if output.StatusCode != http.StatusOK {
		return fmt.Errorf("delete file failed, status code: %d", output.StatusCode)
	}

	return nil
}

func (t *TOS) DeleteFileList(ctx context.Context, keys []string) error {
	objs := make([]tos.ObjectTobeDeleted, 0, len(keys))
	for _, key := range keys {
		objs = append(objs, tos.ObjectTobeDeleted{
			Key: key,
		})
	}
	input := &tos.DeleteMultiObjectsInput{
		Bucket:  t.config.Bucket,
		Objects: objs,
	}

	_, err := t.client.DeleteMultiObjects(ctx, input)
	if err != nil {
		return err
	}

	return nil
}

func (t *TOS) Copy(ctx context.Context, srcKey string, destKey string) (err error) {
	input := &tos.CopyObjectInput{
		Bucket:    t.config.Bucket,
		Key:       destKey,
		SrcBucket: t.config.Bucket,
		SrcKey:    srcKey,
	}

	_, err = t.client.CopyObject(ctx, input)
	if err != nil {
		return err
	}

	return nil
}

func (t *TOS) Move(ctx context.Context, srcKey string, destKey string) (err error) {
	err = t.Copy(ctx, srcKey, destKey)
	if err != nil {
		return err
	}

	err = t.Delete(ctx, srcKey)
	return
}

func (t *TOS) List(ctx context.Context, prefix string, offset int, limit int) (infos []FileInfo, err error) {
	input := &tos.ListObjectsV2Input{
		Bucket: t.config.Bucket,
		ListObjectsInput: tos.ListObjectsInput{
			Prefix:  prefix,
			MaxKeys: offset + limit,
		},
	}
	output, err := t.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, err
	}

	for _, obj := range output.Contents {
		infos = append(infos, FileInfo{
			Bucket:   t.config.Bucket,
			Name:     obj.Key,
			Size:     obj.Size,
			Hash:     obj.ETag,
			MimeType: obj.ObjectType,
		})
	}
	return infos, nil
}

func (t *TOS) GetFileUrl(ctx context.Context, key string) (url string, err error) {
	input := &tos.PreSignedURLInput{
		Bucket:     t.config.Bucket,
		HTTPMethod: http.MethodGet,
		Key:        key,
		Expires:    300,
	}
	output, err := t.client.PreSignedURL(input)
	if err != nil {
		return "", err
	}

	return output.SignedUrl, nil
}

func (t *TOS) Symlink(ctx context.Context, srcKey string, destKey string) (err error) {
	input := &tos.PutSymlinkInput{
		Bucket:              t.config.Bucket,
		Key:                 destKey,
		SymlinkTargetKey:    srcKey,
		SymlinkTargetBucket: t.config.Bucket,
	}

	_, err = t.client.PutSymlink(ctx, input)
	if err != nil {
		return err
	}

	return nil
}
