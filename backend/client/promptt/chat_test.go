package promptt

import (
	"context"
	"fmt"
	"testing"

	"github.com/qianlnk/VirtualHumanStudio/backend/config"
)

func TestModels(t *testing.T) {
	cfg := &config.Promptt{
		BaseURL: "https://aigc-backend.skyengine.com.cn",
		APIKey:  "xiangmuzuzhushou133:ed13cede3dfbd663f330cf54236eaf94",
	}
	p := New(cfg)
	models, err := p.Models()
	fmt.Println(1)
	fmt.Println(models, err)
}

func TestChat(t *testing.T) {
	cfg := &config.Promptt{
		BaseURL: "https://aigc-backend.skyengine.com.cn",
		APIKey:  "xiangmuzuzhushou133:ed13cede3dfbd663f330cf54236eaf94",
	}
	p := New(cfg)
	req := &ChatCompletionRequest{}
	req.Model = "DeepSeek-R1"
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
	fmt.Println(res, err)
}
