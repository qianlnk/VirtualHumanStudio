package promptt

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	resty "github.com/go-resty/resty/v2"
	"github.com/qianlnk/VirtualHumanStudio/backend/utils"
	"github.com/sashabaranov/go-openai"
)

// ChatRequest 聊天请求结构体
type ChatCompletionRequest struct {
	openai.ChatCompletionRequest
	Messages []ChatCompletionMessage `json:"messages"`
}

type ChatCompletionMessage struct {
	openai.ChatCompletionMessage
	MultiContent []ChatMessagePart
}

type ChatMessagePart struct {
	openai.ChatMessagePart
	VideoURL *ChatMessageVideoURL `json:"video_url,omitempty"`
}

type ChatMessageVideoURL struct {
	URL string   `json:"url"`
	Fps *float64 `json:"fps,omitempty"` // 默认1 取值范围 [0.2, 5], 每秒钟从视频中抽取指定数量的图像
}

// ChatResponse 聊天响应结构体
type ChatCompletionResponse struct {
	openai.ChatCompletionResponse

	rawRsp *resty.Response
}

type ChatCompletionStreamResponse struct {
	openai.ChatCompletionStreamResponse
}

func (m ChatCompletionMessage) MarshalJSON() ([]byte, error) {
	if m.Content != "" && m.MultiContent != nil {
		return nil, openai.ErrContentFieldsMisused
	}
	if len(m.MultiContent) > 0 {
		msg := struct {
			Role             string               `json:"role"`
			Content          string               `json:"-"`
			Refusal          string               `json:"refusal,omitempty"`
			MultiContent     []ChatMessagePart    `json:"content,omitempty"`
			Name             string               `json:"name,omitempty"`
			FunctionCall     *openai.FunctionCall `json:"function_call,omitempty"`
			ToolCalls        []openai.ToolCall    `json:"tool_calls,omitempty"`
			ToolCallID       string               `json:"tool_call_id,omitempty"`
			ReasoningContent string               `json:"reasoning_content,omitempty"`
		}{
			m.Role,
			m.Content,
			m.Refusal,
			m.MultiContent,
			m.Name,
			m.FunctionCall,
			m.ToolCalls,
			m.ToolCallID,
			m.ReasoningContent,
		}
		return json.Marshal(msg)
	}

	msg := struct {
		Role             string               `json:"role"`
		Content          string               `json:"content"`
		Refusal          string               `json:"refusal,omitempty"`
		MultiContent     []ChatMessagePart    `json:"-"`
		Name             string               `json:"name,omitempty"`
		FunctionCall     *openai.FunctionCall `json:"function_call,omitempty"`
		ToolCalls        []openai.ToolCall    `json:"tool_calls,omitempty"`
		ToolCallID       string               `json:"tool_call_id,omitempty"`
		ReasoningContent string               `json:"reasoning_content,omitempty"`
	}{
		m.Role,
		m.Content,
		m.Refusal,
		m.MultiContent,
		m.Name,
		m.FunctionCall,
		m.ToolCalls,
		m.ToolCallID,
		m.ReasoningContent,
	}
	return json.Marshal(msg)
}

func (m *ChatCompletionMessage) UnmarshalJSON(bs []byte) error {
	msg := struct {
		Role                     string `json:"role"`
		Content                  string `json:"content"`
		Refusal                  string `json:"refusal,omitempty"`
		MultiContent             []ChatMessagePart
		Name                     string               `json:"name,omitempty"`
		FunctionCall             *openai.FunctionCall `json:"function_call,omitempty"`
		ToolCalls                []openai.ToolCall    `json:"tool_calls,omitempty"`
		ToolCallID               string               `json:"tool_call_id,omitempty"`
		ReasoningContent         string               `json:"reasoning_content,omitempty"`
		Signature                string               `json:"signature,omitempty"`
		RedactedReasoningContent string               `json:"redacted_reasoning_content,omitempty"`
	}{}
	if err := json.Unmarshal(bs, &msg); err == nil {
		*m = ChatCompletionMessage{}
		m.Role = msg.Role
		m.Content = msg.Content
		m.Refusal = msg.Refusal
		m.MultiContent = msg.MultiContent
		m.Name = msg.Name
		m.FunctionCall = msg.FunctionCall
		m.ToolCalls = msg.ToolCalls
		m.ToolCallID = msg.ToolCallID
		m.ReasoningContent = msg.ReasoningContent
		return nil
	}

	multiMsg := struct {
		Role                     string               `json:"role"`
		Content                  string               `json:"-"`
		Refusal                  string               `json:"refusal,omitempty"`
		MultiContent             []ChatMessagePart    `json:"content"`
		Name                     string               `json:"name,omitempty"`
		FunctionCall             *openai.FunctionCall `json:"function_call,omitempty"`
		ToolCalls                []openai.ToolCall    `json:"tool_calls,omitempty"`
		ToolCallID               string               `json:"tool_call_id,omitempty"`
		ReasoningContent         string               `json:"reasoning_content,omitempty"`
		Signature                string               `json:"signature,omitempty"`
		RedactedReasoningContent string               `json:"redacted_reasoning_content,omitempty"`
	}{}
	if err := json.Unmarshal(bs, &multiMsg); err != nil {
		return err
	}
	*m = ChatCompletionMessage{}
	m.Role = multiMsg.Role
	m.Content = multiMsg.Content
	m.Refusal = multiMsg.Refusal
	m.MultiContent = multiMsg.MultiContent
	m.Name = multiMsg.Name
	m.FunctionCall = multiMsg.FunctionCall
	m.ToolCalls = multiMsg.ToolCalls
	m.ToolCallID = multiMsg.ToolCallID
	m.ReasoningContent = multiMsg.ReasoningContent

	return nil
}

// DoChat 调用大语言模型进行聊天
func (p *Promptt) DoChat(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	fmt.Println(p.cfg.BaseURL, "apikey:", p.cfg.APIKey)
	chatRes := ChatCompletionResponse{}
	resp, err := p.cli.R().SetHeader("Authorization", "Bearer "+p.cfg.APIKey).
		SetBody(req).
		SetResult(&chatRes).
		SetDoNotParseResponse(req.Stream).
		Post("/eliza/v2/chat/completions")
	if err != nil {
		return nil, fmt.Errorf("创建聊天请求失败: %v", err)
	}
	if resp.StatusCode() != http.StatusOK {
		fmt.Println(resp.StatusCode(), "req: %s", utils.ToJSONString(req))
		return nil, fmt.Errorf("创建聊天请求失败: %v", string(resp.Body()))
	}

	chatRes.rawRsp = resp

	return &chatRes, nil
}

func (r *ChatCompletionResponse) GetCompletionCh() (<-chan ChatCompletionStreamResponse, error) {
	c := make(chan ChatCompletionStreamResponse)
	prefix := "data:"
	prefixLen := len(prefix)
	stopLine := "data: [DONE]"
	go func() {
		defer close(c)
		body := r.rawRsp.RawBody()
		defer body.Close()
		reader := bufio.NewReader(body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("read line failed: %v", err)
				return
			}
			if strings.HasPrefix(line, prefix) {
				line = line[prefixLen:]
				if line == stopLine {
					break
				}
				var streamResp ChatCompletionStreamResponse
				if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
					fmt.Printf("unmarshal stream response failed: %v", err)
					continue
				}
				c <- streamResp

				if streamResp.Usage != nil {
					r.Usage = *streamResp.Usage
				}
			}
		}
	}()

	return c, nil
}

type ModelInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Service   string `json:"service"`
	ModelType string `json:"model_type"`
	Kind      string `json:"kind"`
	Created   int    `json:"created"`
	OwnedBy   string `json:"owned_by"`
	Caps      struct {
		Chat   bool `json:"chat"`
		Vision bool `json:"vision"`
		Tts    bool `json:"tts"`
		Asr    bool `json:"asr"`
		Image  bool `json:"image"`
	} `json:"caps"`
	Params struct {
		VoiceList []any `json:"voice_list"`
	} `json:"params"`
	Characteristics []string `json:"characteristics"`
}

type GetModelsResponse struct {
	Data []ModelInfo `json:"data"`
}

func (p *Promptt) Models() (*GetModelsResponse, error) {
	if p.modelListExpiredAt.After(time.Now()) && p.modelList != nil {
		return p.modelList, nil
	}

	resp, err := p.cli.R().SetHeader("Authorization", "Bearer "+p.cfg.APIKey).
		Get("/eliza/v1/models")
	if err != nil {
		return nil, fmt.Errorf("获取模型列表失败: %v", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("获取模型列表失败: %v", resp.String())
	}

	fmt.Println(resp.String())

	var res = new(GetModelsResponse)
	if err := json.Unmarshal(resp.Body(), res); err != nil {
		return nil, fmt.Errorf("解析模型列表失败: %v", err)
	}

	// 排序
	sort.SliceStable(res.Data, func(i, j int) bool {
		return res.Data[i].Service < res.Data[j].Service
	})

	p.modelList = res
	p.modelListExpiredAt = time.Now().Add(time.Minute * 10)

	return res, nil
}

const (
	ModelTypeLLM     = "TXT_TO_TXT_MODEL"
	ModelTypeDrawing = "TXT_TO_IMG_MODEL"
	ModelTypeTTS     = "TTS_MODEL"
	ModelTypeASR     = "ASR_MODEL"
)
