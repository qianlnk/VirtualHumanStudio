package storages

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var emptyKey = errors.New("key is empty")

type FSConfig struct {
	BasePath string `yaml:"base_path" json:"base_path"`
	BaseURL  string `yaml:"base_url" json:"base_url"`
}

type FS struct {
	cfg FSConfig
}

func NewFS(cfg FSConfig) (*FS, error) {
	return &FS{
		cfg: cfg,
	}, nil
}

func (f *FS) Init() error {
	return nil
}

func (f *FS) UploadToken(ctx context.Context, policy *PutPolicy) (method string, host string, header map[string]string, form map[string]string, usage string, err error) {
	return "", "", nil, nil, "", nil
}

// 保存文件
func (f *FS) SaveFile(ctx context.Context, key string, reader io.Reader) (err error) {
	if key == "" {
		return emptyKey
	}

	filename := path.Join(f.cfg.BasePath, key)
	paths := strings.Split(filename, "/")
	dir := strings.Join(paths[:len(paths)-1], "/")
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return err
	}

	return nil
}

// 打开文件
func (f *FS) OpenFile(ctx context.Context, key string) (reader io.ReadCloser, err error) {
	if key == "" {
		return nil, emptyKey
	}

	filename := path.Join(f.cfg.BasePath, key)
	return os.Open(filename)
}

// 获取文件
func (f *FS) FetchFile(ctx context.Context, key string) (data []byte, err error) {
	if key == "" {
		return nil, emptyKey
	}

	filename := path.Join(f.cfg.BasePath, key)
	return os.ReadFile(filename)
}

// 文件状态
func (f *FS) StatFile(ctx context.Context, key string) (info FileInfo, err error) {
	if key == "" {
		return FileInfo{}, emptyKey
	}

	filename := path.Join(f.cfg.BasePath, key)
	fi, err := os.Stat(filename)
	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{
		Name:      fi.Name(),
		Size:      fi.Size(),
		UpdatedAt: fi.ModTime().UnixMilli(),
	}, nil
}

// 删除文件
func (f *FS) Delete(ctx context.Context, key string) (err error) {
	if key == "" {
		return nil
	}

	filename := path.Join(f.cfg.BasePath, key)
	return os.RemoveAll(filename)
}

// 复制文件
func (f *FS) Copy(ctx context.Context, srcKey string, destKey string) (err error) {
	if srcKey == "" || destKey == "" {
		return emptyKey
	}

	destFilename := path.Join(f.cfg.BasePath, destKey)
	srcFilename := path.Join(f.cfg.BasePath, srcKey)
	paths := strings.Split(destFilename, "/")
	dir := strings.Join(paths[:len(paths)-1], "/")
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	srcFile, err := os.Open(srcFilename)
	if err != nil {
		return err
	}

	destFile, err := os.Create(destFilename)
	if err != nil {
		return err
	}

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}

// 移动文件
func (f *FS) Move(ctx context.Context, srcKey string, destKey string) (err error) {
	if srcKey == "" || destKey == "" {
		return emptyKey
	}

	destFilename := path.Join(f.cfg.BasePath, destKey)
	srcFilename := path.Join(f.cfg.BasePath, srcKey)
	paths := strings.Split(destFilename, "/")
	dir := strings.Join(paths[:len(paths)-1], "/")
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	return os.Rename(srcFilename, destFilename)
}

// 遍历文件夹及其子目录中的所有文件
func (f *FS) List(ctx context.Context, prefix string, offset int, limit int) (infos []FileInfo, err error) {
	count := 0
	prefix = path.Join(f.cfg.BasePath, prefix)
	basePath := f.cfg.BasePath
	info, err := os.Stat(prefix)
	if err == nil {
		if !info.IsDir() {
			infos = append(infos, FileInfo{
				Name:      info.Name(),
				Ext:       filepath.Ext(info.Name()),
				Size:      info.Size(),
				Path:      prefix,
				UpdatedAt: info.ModTime().UnixMilli(),
			})

			return
		} else {
			basePath = prefix
		}
	}

	err = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			count++
			if count <= offset {
				return nil
			}

			if strings.HasPrefix(path, prefix) {
				infos = append(infos, FileInfo{
					Name:      info.Name(),
					Ext:       filepath.Ext(info.Name()),
					Size:      info.Size(),
					Path:      path,
					UpdatedAt: info.ModTime().UnixMilli(),
				})
			}
		}

		if len(infos) >= limit && limit != 0 {
			return errors.New("break")
		}

		return nil
	})

	if err != nil {
		if err.Error() == "break" {
			err = nil
		}

		return
	}

	return
}

// 创建软链接
func (f *FS) Symlink(ctx context.Context, srcKey string, destKey string) (err error) {
	if srcKey == "" || destKey == "" {
		return emptyKey
	}

	sourceDir := path.Join(f.cfg.BasePath, srcKey)
	linkPath := path.Join(f.cfg.BasePath, destKey)

	//先删除软链接
	os.Remove(linkPath)

	err = os.Symlink(sourceDir, linkPath)
	if err != nil {
		return
	}

	return
}

// 获取文件url
func (f *FS) GetFileUrl(ctx context.Context, key string) (url string, err error) {
	if key == "" {
		return "", emptyKey
	}

	return fmt.Sprintf("%s/api/file/view?path=%s", f.cfg.BaseURL, key), nil
}
