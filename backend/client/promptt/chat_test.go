package promptt

import (
	"context"
	"fmt"
	"testing"

	"github.com/qianlnk/VirtualHumanStudio/backend/config"
	"github.com/qianlnk/VirtualHumanStudio/backend/utils"
)

func TestModels(t *testing.T) {
	cfg := &config.Promptt{
		BaseURL: "https://aigc-backend.skyengine.com.cn",
		APIKey:  "xxxxx",
	}
	p := New(cfg)
	models, err := p.Models()
	fmt.Println(1)
	fmt.Println(models, err)
}

func TestChat(t *testing.T) {
	cfg := &config.Promptt{
		BaseURL: "https://aigc-backend.skyengine.com.cn",
		APIKey:  "xxxxx",
	}
	p := New(cfg)
	req := &ChatCompletionRequest{}
	req.Model = "claude-opus-4.1-20250805-thinking"
	req.Stream = false
	msg := ChatCompletionMessage{}
	msg.Role = "user"
	part := ChatMessagePart{}
	part.Type = "text"
	part.Text = "你好"

	msg.MultiContent = append(msg.MultiContent, part)
	req.Messages = append(req.Messages, msg)
	res, err := p.DoChat(context.Background(), req)
	fmt.Println(1)
	fmt.Println(utils.ToJSONString(res), err)
}
