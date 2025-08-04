package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/qianlnk/VirtualHumanStudio/backend/client/promptt"
	"github.com/qianlnk/VirtualHumanStudio/backend/models"
	"github.com/sashabaranov/go-openai"
)

// ChatController 聊天控制器
type ChatController struct {
	chatRepo models.ChatRepository
	promptt  *promptt.Promptt
}

// NewChatController 创建新的聊天控制器
func NewChatController(chatRepo models.ChatRepository, prompttClient *promptt.Promptt) *ChatController {
	return &ChatController{
		chatRepo: chatRepo,
		promptt:  prompttClient,
	}
}

// CreateSession 创建新的聊天会话
func (cc *ChatController) CreateSession(c *gin.Context) {
	var req struct {
		Title string `json:"title"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	sessionID := generateSessionID()
	session := &models.ChatSession{
		ID:           sessionID,
		Title:        req.Title,
		UserID:       userID.(uint),
		MessageCount: 0,
	}

	if err := cc.chatRepo.CreateSession(session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建会话失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"session": session,
	})
}

// GetSessions 获取聊天会话列表
func (cc *ChatController) GetSessions(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	sessions, err := cc.chatRepo.GetSessionsByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取会话列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"sessions": sessions,
	})
}

// GetSessionMessages 获取会话消息历史
func (cc *ChatController) GetSessionMessages(c *gin.Context) {
	sessionID := c.Param("sessionID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话ID不能为空"})
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 验证会话是否属于当前用户
	session, err := cc.chatRepo.GetSessionByID(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
		return
	}

	if session.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问此会话"})
		return
	}

	messages, err := cc.chatRepo.GetMessagesBySessionID(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取消息历史失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"messages": messages,
	})
}

type SendMessageRequest struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
	Model     string `json:"model"`
	ImageURL  string `json:"image_url,omitempty"`
	VideoURL  string `json:"video_url,omitempty"`
	Stream    bool   `json:"stream,omitempty"` // 是否使用流式返回
}

// SendMessage 发送消息并获取AI回复
func (cc *ChatController) SendMessage(c *gin.Context) {
	var req SendMessageRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	if req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "消息内容不能为空"})
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 验证会话是否属于当前用户
	session, err := cc.chatRepo.GetSessionByID(req.SessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
		return
	}

	if session.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问此会话"})
		return
	}

	// 保存用户消息
	userMessage := &models.ChatMessage{
		ID:        generateMessageID(),
		SessionID: req.SessionID,
		UserID:    userID.(uint),
		Role:      "user",
		Content:   req.Message,
		ImageURL:  req.ImageURL,
		VideoURL:  req.VideoURL,
	}

	if err := cc.chatRepo.CreateMessage(userMessage); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存用户消息失败"})
		return
	}

	// 如果请求流式返回，使用SSE
	if req.Stream {
		cc.handleStreamResponse(c, req, userMessage, session)
		return
	}

	// 非流式返回（原有逻辑）
	cc.handleNormalResponse(c, req, userMessage, session)
}

// handleStreamResponse 处理流式响应
func (cc *ChatController) handleStreamResponse(c *gin.Context, req SendMessageRequest, userMessage *models.ChatMessage, session *models.ChatSession) {
	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")
	c.Header("X-Accel-Buffering", "no") // 禁用 nginx 缓冲

	// 确保立即发送响应头
	c.Writer.Flush()

	// 获取用户ID
	userID, _ := c.Get("user_id")

	modelName := req.Model
	// 创建AI消息记录
	aiMessage := &models.ChatMessage{
		ID:        generateMessageID(),
		SessionID: req.SessionID,
		UserID:    userID.(uint),
		Role:      "assistant",
		Content:   "", // 初始为空，后续会更新
		Model:     modelName,
	}

	if err := cc.chatRepo.CreateMessage(aiMessage); err != nil {
		c.SSEvent("error", gin.H{"error": "保存AI消息失败", "model": modelName})
		return
	}

	// 发送开始事件
	c.SSEvent("start", gin.H{
		"model":      modelName,
		"message_id": aiMessage.ID,
	})
	c.Writer.Flush() // 立即发送

	// 调用流式AI接口（传入会话ID）
	responseText, imageURL, err := cc.generateStreamAIResponse(c, req.Message, modelName, req.ImageURL, aiMessage.ID, req.SessionID)
	if err != nil {
		// 发送错误事件
		c.SSEvent("error", gin.H{
			"error":      fmt.Sprintf("抱歉，%s模型暂时无法响应，请稍后重试， %s", modelName, err),
			"model":      modelName,
			"message_id": aiMessage.ID,
		})
		c.Writer.Flush() // 立即发送
		return
	}

	// 更新AI消息内容
	aiMessage.Content = responseText
	aiMessage.ImageURL = imageURL
	if err := cc.chatRepo.UpdateMessage(aiMessage); err != nil {
		c.SSEvent("error", gin.H{"error": "更新AI消息失败", "model": modelName})
	}

	// 发送完成事件
	c.SSEvent("complete", gin.H{
		"model":      modelName,
		"message_id": aiMessage.ID,
		"content":    responseText,
		"image_url":  imageURL,
	})
	c.Writer.Flush() // 立即发送

	// 更新会话信息
	session.MessageCount += 2
	session.UpdatedAt = time.Now()
	if err := cc.chatRepo.UpdateSession(session); err != nil {
		c.SSEvent("error", gin.H{"error": "更新会话失败"})
		c.Writer.Flush() // 立即发送
	}

	// 发送结束事件
	c.SSEvent("end", gin.H{"success": true})
	c.Writer.Flush() // 立即发送
}

// handleNormalResponse 处理普通响应
func (cc *ChatController) handleNormalResponse(c *gin.Context, req SendMessageRequest, userMessage *models.ChatMessage, session *models.ChatSession) {
	// 获取用户ID
	userID, _ := c.Get("user_id")

	// 生成AI回复（多模型）
	var aiMessages []*models.ChatMessage
	modelName := req.Model
	// 调用AI接口生成回复（传入会话ID，以便获取历史消息）
	aiResponse, imageURL, err := cc.generateAIResponse(req.Message, modelName, req.ImageURL, req.SessionID)
	if err != nil {
		// 如果AI调用失败，创建一个错误消息
		aiResponse = fmt.Sprintf("抱歉，%s模型暂时无法响应，请稍后重试。", modelName)
		imageURL = ""
	}

	aiMessage := &models.ChatMessage{
		ID:        generateMessageID(),
		SessionID: req.SessionID,
		UserID:    userID.(uint),
		Role:      "assistant",
		Content:   aiResponse,
		Model:     modelName,
		ImageURL:  imageURL,
	}

	if err := cc.chatRepo.CreateMessage(aiMessage); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存AI消息失败"})
		return
	}

	aiMessages = append(aiMessages, aiMessage)

	// 更新会话信息
	session.MessageCount += 2
	session.UpdatedAt = time.Now()
	if err := cc.chatRepo.UpdateSession(session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新会话失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"user_message": userMessage,
		"ai_messages":  aiMessages,
	})
}

// generateAIResponse 生成AI回复
func (cc *ChatController) generateAIResponse(userMessage, modelName, imageURL, sessionID string) (string, string, error) {
	ctx := context.Background()

	// 根据模型类型选择不同的处理方式
	switch modelName {
	case "DALL-E 3", "Midjourney", "Stable Diffusion":
		// 图像生成模型
		return cc.generateImage(ctx, userMessage, modelName, imageURL)
	case "Claude 3.5 Sonnet", "GPT-4", "GPT-3.5-turbo":
		// 文本对话模型
		return cc.generateText(ctx, userMessage, modelName, sessionID)
	default:
		// 默认使用文本对话
		return cc.generateText(ctx, userMessage, modelName, sessionID)
	}
}

// generateImage 生成图像
func (cc *ChatController) generateImage(ctx context.Context, prompt, modelName, imageURL string) (string, string, error) {
	// 构建绘画请求
	paintReq := &promptt.PaintRequest{
		Model:  modelName,
		Prompt: prompt,
		Size:   "1024x1024", // 默认尺寸
		N:      1,           // 生成1张图片
	}

	// 如果有参考图片，添加到请求中
	if imageURL != "" {
		paintReq.ImageURLs = []string{imageURL}
	}

	// 调用绘画服务
	paintResp, err := cc.promptt.DoPaint(ctx, paintReq)
	if err != nil {
		return "", "", fmt.Errorf("调用绘画服务失败: %v", err)
	}

	// 生成回复文本
	responseText := fmt.Sprintf("我已经为您生成了图像，展现了您描述的场景：%s", prompt)

	// 返回生成的图像URL
	var generatedImageURL string
	if len(paintResp.Images) > 0 {
		generatedImageURL = paintResp.Images[0].URL
	}

	return responseText, generatedImageURL, nil
}

// generateText 生成文本回复
func (cc *ChatController) generateText(ctx context.Context, userMessage, modelName, sessionID string) (string, string, error) {
	// 构建聊天请求
	var chatReq *promptt.ChatCompletionRequest
	var promptMessages []promptt.ChatCompletionMessage

	// 如果提供了会话ID，则获取历史消息
	if sessionID != "" {
		messages, err := cc.chatRepo.GetMessagesBySessionID(sessionID)
		if err == nil && len(messages) > 0 {
			// 如果成功获取历史消息，构建带有历史记录的请求
			for _, msg := range messages {
				promptMessages = append(promptMessages, promptt.ChatCompletionMessage{
					ChatCompletionMessage: openai.ChatCompletionMessage{
						Role:    msg.Role,
						Content: msg.Content,
					},
				})
			}
		}
	}

	// 如果没有历史消息或无法获取历史消息，使用单条消息请求
	if len(promptMessages) == 0 {
		promptMessages = append(promptMessages, promptt.ChatCompletionMessage{
			ChatCompletionMessage: openai.ChatCompletionMessage{
				Role:    "user",
				Content: userMessage,
			},
		})
	} else {
		// 如果最后一条消息不是当前用户消息，则添加
		lastMsg := promptMessages[len(promptMessages)-1]
		if lastMsg.Role != "user" || lastMsg.Content != userMessage {
			promptMessages = append(promptMessages, promptt.ChatCompletionMessage{
				ChatCompletionMessage: openai.ChatCompletionMessage{
					Role:    "user",
					Content: userMessage,
				},
			})
		}
	}

	// 构建聊天请求
	chatReq = &promptt.ChatCompletionRequest{
		ChatCompletionRequest: openai.ChatCompletionRequest{
			Model: modelName,
		},
		Messages: promptMessages,
	}

	// 调用聊天服务
	chatResp, err := cc.promptt.DoChat(ctx, chatReq)
	if err != nil {
		return "", "", fmt.Errorf("调用聊天服务失败: %v", err)
	}

	// 提取回复内容
	var responseText string
	if len(chatResp.Choices) > 0 {
		responseText = chatResp.Choices[0].Message.Content
	} else {
		responseText = "抱歉，我暂时无法理解您的请求，请稍后重试。"
	}

	return responseText, "", nil
}

// generateStreamAIResponse 生成流式AI回复
func (cc *ChatController) generateStreamAIResponse(c *gin.Context, userMessage, modelName, imageURL, messageID, sessionID string) (string, string, error) {
	ctx := context.Background()

	// 根据模型类型选择不同的处理方式
	switch modelName {
	case "DALL-E 3", "Midjourney", "Stable Diffusion":
		// 图像生成模型不支持流式返回，使用普通方式
		return cc.generateImage(ctx, userMessage, modelName, imageURL)
	case "Claude 3.5 Sonnet", "GPT-4", "GPT-3.5-turbo":
		// 文本对话模型支持流式返回
		return cc.generateStreamText(c, ctx, userMessage, modelName, messageID, sessionID)
	default:
		// 默认使用流式文本对话
		return cc.generateStreamText(c, ctx, userMessage, modelName, messageID, sessionID)
	}
}

// generateStreamText 生成流式文本回复
func (cc *ChatController) generateStreamText(c *gin.Context, ctx context.Context, userMessage, modelName, messageID, sessionID string) (string, string, error) {
	// 构建聊天请求
	var promptMessages []promptt.ChatCompletionMessage

	// 获取会话的历史消息
	if sessionID != "" {
		messages, err := cc.chatRepo.GetMessagesBySessionID(sessionID)
		if err == nil && len(messages) > 0 {
			// 如果成功获取历史消息，则添加到请求中
			for _, msg := range messages {
				promptMessages = append(promptMessages, promptt.ChatCompletionMessage{
					ChatCompletionMessage: openai.ChatCompletionMessage{
						Role:    msg.Role,
						Content: msg.Content,
					},
				})
			}
		}
	}

	// 无论是否有历史消息，都添加当前用户消息
	promptMessages = append(promptMessages, promptt.ChatCompletionMessage{
		ChatCompletionMessage: openai.ChatCompletionMessage{
			Role:    "user",
			Content: userMessage,
		},
	})

	// 构建聊天请求
	chatReq := &promptt.ChatCompletionRequest{
		ChatCompletionRequest: openai.ChatCompletionRequest{
			Model:  modelName,
			Stream: true, // 启用流式返回
		},
		Messages: promptMessages,
	}

	// 调用聊天服务
	chatResp, err := cc.promptt.DoChat(ctx, chatReq)
	if err != nil {
		return "", "", fmt.Errorf("调用聊天服务失败: %v", err)
	}

	// 获取流式响应通道
	streamCh, err := chatResp.GetCompletionCh()
	if err != nil {
		return "", "", fmt.Errorf("获取流式响应失败: %v", err)
	}

	var fullResponse string
	// 处理流式响应
	for chunk := range streamCh {
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta
			if delta.Content != "" {
				fullResponse += delta.Content
				// 发送流式数据到前端
				c.SSEvent("chunk", gin.H{
					"message_id": messageID,
					"model":      modelName,
					"content":    delta.Content,
					"delta":      delta,
				})
				c.Writer.Flush() // 立即发送，确保数据实时到达客户端
			}
		}
	}

	return fullResponse, "", nil
}

// GetAvailableModels 获取可用的AI模型列表
func (cc *ChatController) GetAvailableModels(c *gin.Context) {

	res, err := cc.promptt.Models()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取模型列表失败"})
		return
	}

	var ms []models.ModelInfo
	for _, m := range res.Data {
		if m.ModelType == promptt.ModelTypeLLM {
			ms = append(ms, models.ModelInfo{
				ID:   m.ID,
				Name: m.Service,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"models":  ms,
	})
}

// DeleteSession 删除聊天会话
func (cc *ChatController) DeleteSession(c *gin.Context) {
	sessionID := c.Param("sessionID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话ID不能为空"})
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 验证会话是否属于当前用户
	session, err := cc.chatRepo.GetSessionByID(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
		return
	}

	if session.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除此会话"})
		return
	}

	if err := cc.chatRepo.DeleteSession(sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除会话失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "会话删除成功",
	})
}

// 辅助函数
func generateSessionID() string {
	return uuid.NewString()
}

func generateMessageID() string {
	return uuid.NewString()
}
