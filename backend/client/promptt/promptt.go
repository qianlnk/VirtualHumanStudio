package promptt

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	resty "github.com/go-resty/resty/v2"
	"github.com/qianlnk/VirtualHumanStudio/backend/config"
	openai "github.com/sashabaranov/go-openai"
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

		// 创建本地文件
		f, err := os.Create(savePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create local file: %v", err)
		}
		defer f.Close()

		// 将响应内容写入文件
		if _, err := io.Copy(f, r.RawBody()); err != nil {
			return nil, fmt.Errorf("failed to save audio file: %v", err)
		}
	}

	return &resp, nil
}

// ChatRequest 聊天请求结构体
type ChatRequest struct {
	openai.ChatCompletionRequest
}

// ChatResponse 聊天响应结构体
type ChatResponse struct {
	ID      string                              `json:"id"`
	Choices []openai.ChatCompletionStreamChoice `json:"choices"`
	Usage   openai.Usage                        `json:"usage"`
}

// DoChat 调用大语言模型进行聊天
func (p *Promptt) DoChat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// 创建OpenAI客户端
	client := openai.NewClient(p.cfg.APIKey)

	// 构建请求
	chatReq := openai.ChatCompletionRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		Stream:      req.Stream,
	}

	// 根据是否为流式请求选择不同的处理方式
	if req.Stream {
		// 创建流式聊天完成请求
		stream, err := client.CreateChatCompletionStream(ctx, chatReq)
		if err != nil {
			return nil, fmt.Errorf("创建流式聊天请求失败: %v", err)
		}
		defer stream.Close()

		// 处理流式响应
		var response ChatResponse
		for {
			resp, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				return nil, fmt.Errorf("接收流式响应失败: %v", err)
			}
			// 更新响应
			response.ID = resp.ID
			response.Choices = append(response.Choices, openai.ChatCompletionStreamChoice{
				Index:        resp.Choices[0].Index,
				Delta:        resp.Choices[0].Delta,
				FinishReason: resp.Choices[0].FinishReason,
			})
		}
		return &response, nil
	}

	// 非流式请求
	resp, err := client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("创建聊天请求失败: %v", err)
	}

	// 转换响应格式
	return &ChatResponse{
		ID: resp.ID,
		Choices: []openai.ChatCompletionStreamChoice{
			{
				Index: 0,
				Delta: openai.ChatCompletionStreamChoiceDelta{
					Content: resp.Choices[0].Message.Content,
				},
				FinishReason: resp.Choices[0].FinishReason,
			},
		},
		Usage: resp.Usage,
	}, nil
}
