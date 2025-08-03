package promptt

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	resty "github.com/go-resty/resty/v2"
	"github.com/qianlnk/VirtualHumanStudio/backend/config"
	"github.com/qianlnk/VirtualHumanStudio/backend/storages"
	"github.com/qianlnk/VirtualHumanStudio/backend/utils"
)

type TTSRequest struct {
	Model      string                 `json:"model"`
	Input      string                 `json:"input"`
	Voice      string                 `json:"voice"`
	ModelExtra map[string]interface{} `json:"model_extra,omitempty"`
}

type TTSResponse struct {
	RequestID  string `json:"request_id"`
	URL        string `json:"url"`
	DurationMs int    `json:"duration_ms"`
}

type Promptt struct {
	cfg *config.Promptt
	cli *resty.Client

	modelList          *GetModelsResponse
	modelListExpiredAt time.Time
}

func New(cfg *config.Promptt) *Promptt {
	return &Promptt{
		cfg: cfg,
		cli: resty.New().SetBaseURL(cfg.BaseURL),
	}
}

// DoTTS 调用TTS服务将文本转换为语音
func (p *Promptt) DoTTS(req *TTSRequest, savePath string) (*TTSResponse, error) {
	// 发送TTS请求
	var resp TTSResponse
	r, err := p.cli.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", p.cfg.APIKey)).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(&resp).
		Post(p.cfg.TTSPath)

	if err != nil {
		return nil, fmt.Errorf("failed to call TTS service: %v", err)
	}

	if r.IsError() {
		return nil, fmt.Errorf("TTS service returned error: %s", r.String())
	}

	// 下载音频文件
	if savePath != "" {
		// 确保目录存在
		if err := os.MkdirAll(filepath.Dir(savePath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %v", err)
		}

		// 下载音频文件
		r, err := p.cli.R().SetDoNotParseResponse(true).Get(resp.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to download audio file: %v", err)
		}
		defer r.RawBody().Close()

		err = storages.Client.SaveFile(context.Background(), savePath, r.RawBody())
		if err != nil {
			return nil, fmt.Errorf("failed to save audio file: %v", err)
		}
	}

	return &resp, nil
}

type PaintRequest struct {
	Model     string   `json:"model,omitempty"`
	ImageURLs []string `json:"image_urls,omitempty"`
	Mask      []string `json:"mask,omitempty"`
	Prompt    string   `json:"prompt,omitempty"`
	Size      string   `json:"size,omitempty"`
	N         uint32   `json:"n"`
	Height    uint32   `json:"height,omitempty"`
	Width     uint32   `json:"width,omitempty"`
	BatchSize uint32   `json:"batch_size,omitempty"`
}

type PaintResponse struct {
	Images []Image `json:"images"`
}

type Image struct {
	URL    string `json:"url"`
	Path   string `json:"path"`
	Type   string `json:"type"`
	Censor bool   `json:"censor"`
}

func (p *Promptt) DoPaint(ctx context.Context, req *PaintRequest) (*PaintResponse, error) {
	fmt.Println("DoPaint", utils.ToJSONString(req))
	var resp PaintResponse
	r, err := p.cli.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", p.cfg.APIKey)).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(&resp).
		Post(p.cfg.PaintPath)

	if err != nil {
		return nil, fmt.Errorf("failed to call Paint service: %v", err)
	}

	if r.IsError() {
		return nil, fmt.Errorf("Paint service returned error: %s", r.String())
	}

	fmt.Println("DoPaint", utils.ToJSONString(resp))
	return &resp, nil
}
