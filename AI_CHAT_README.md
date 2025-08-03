# AI绘画聊天系统

这是一个简洁的AI绘画聊天页面，支持多模型对比、历史记录存储等功能。系统使用标准的OpenAI协议与AI服务进行通信。

## 功能特性

### 🎨 核心功能
- **多模型支持**: 同时选择多个AI模型进行绘画创作
- **实时对比**: 同一会话中多个模型同时回复，方便对比效果
- **历史记录**: 完整的聊天历史记录保存和查看
- **会话管理**: 创建、删除、切换不同的绘画会话
- **OpenAI协议**: 使用标准OpenAI协议，支持多种AI服务

### 🖼️ 绘画功能
- **文本生成图像**: 通过文字描述生成图像
- **图像上传**: 支持上传参考图片
- **多风格支持**: 支持多种绘画风格和艺术效果
- **图像编辑**: 支持基于参考图像的编辑和变体生成

### 💾 数据存储
- **会话存储**: 自动保存所有聊天会话
- **消息历史**: 完整的对话记录保存
- **用户隔离**: 每个用户只能访问自己的会话

## 技术架构

### 后端 (Go + Gin)
```
backend/
├── controllers/
│   └── chat_controller.go    # 聊天控制器
├── models/
│   └── chat.go              # 数据模型
├── routes/
│   └── chat_routes.go       # API路由
├── client/
│   └── promptt/             # OpenAI协议客户端
│       ├── promptt.go       # 客户端主文件
│       └── chat.go          # 聊天接口
└── middleware/
    └── auth.go              # 认证中间件
```

### 前端 (Vue.js + Element UI)
```
frontend/src/
├── views/
│   └── AIChat.vue           # 主聊天页面
├── api/
│   └── chat.js              # API接口
└── router/
    └── index.js             # 路由配置
```

## OpenAI协议集成

### 支持的模型类型

#### 图像生成模型
- **DALL-E 3**: OpenAI最新的图像生成模型
- **Midjourney**: 专注于艺术创作的AI绘画模型
- **Stable Diffusion**: 开源的图像生成模型

#### 文本对话模型
- **Claude 3.5 Sonnet**: Anthropic的智能模型
- **GPT-4**: OpenAI的GPT-4模型
- **GPT-3.5-turbo**: OpenAI的GPT-3.5模型

### API调用流程

1. **文本对话**:
   ```go
   chatReq := &promptt.ChatCompletionRequest{
       ChatCompletionRequest: openai.ChatCompletionRequest{
           Model: "gpt-4",
       },
       Messages: []promptt.ChatCompletionMessage{
           {
               ChatCompletionMessage: openai.ChatCompletionMessage{
                   Role:    "user",
                   Content: userMessage,
               },
           },
       },
   }
   ```

2. **图像生成**:
   ```go
   paintReq := &promptt.PaintRequest{
       Model:  "dall-e-3",
       Prompt: "画一只可爱的小猫",
       Size:   "1024x1024",
       N:      1,
   }
   ```

## 数据库设计

### 表结构
1. **chat_sessions**: 聊天会话表
   - id: 会话ID
   - title: 会话标题
   - user_id: 用户ID
   - message_count: 消息数量
   - created_at: 创建时间
   - updated_at: 更新时间

2. **chat_messages**: 聊天消息表
   - id: 消息ID
   - session_id: 会话ID
   - user_id: 用户ID
   - role: 角色 (user/assistant)
   - content: 消息内容
   - model: AI模型名称
   - image_url: 图像URL
   - created_at: 创建时间

3. **model_infos**: AI模型信息表
   - id: 模型ID
   - name: 模型名称
   - description: 模型描述
   - context: 上下文信息
   - input_price: 输入价格
   - output_price: 输出价格
   - is_active: 是否激活

## API接口

### 会话管理
- `POST /api/chat/sessions` - 创建新会话
- `GET /api/chat/sessions` - 获取会话列表
- `GET /api/chat/sessions/:sessionID/messages` - 获取会话消息
- `DELETE /api/chat/sessions/:sessionID` - 删除会话

### 消息处理
- `POST /api/chat/messages` - 发送消息并获取AI回复

### 模型信息
- `GET /api/chat/models` - 获取可用模型列表

## 使用说明

### 1. 创建新会话
点击顶部的"新会话"按钮，输入会话标题即可创建新的绘画会话。

### 2. 选择AI模型
在模型选择区域，可以选择一个或多个AI模型：
- **DALL-E 3**: OpenAI最新的图像生成模型
- **Midjourney**: 专注于艺术创作的AI绘画模型
- **Stable Diffusion**: 开源的图像生成模型
- **Claude 3.5 Sonnet**: Anthropic的智能模型

### 3. 发送绘画请求
在底部输入框中描述您想要生成的图像，支持：
- 详细的场景描述
- 艺术风格要求
- 颜色和构图要求
- 参考图片上传

### 4. 查看多模型结果
系统会同时调用选中的多个AI模型，在聊天界面中显示每个模型的回复和生成的图像，方便对比不同模型的效果。

### 5. 管理历史记录
- 左侧会话列表显示所有历史会话
- 点击会话可以查看完整的对话记录
- 支持删除不需要的会话

## 界面设计

### 设计理念
参考了现代聊天应用的简洁设计，采用：
- **清晰的布局**: 左侧会话列表 + 右侧聊天区域
- **直观的操作**: 简单的按钮和交互设计
- **响应式设计**: 适配不同屏幕尺寸
- **现代化UI**: 使用Element UI组件库

### 主要界面元素
1. **顶部导航栏**: 新会话、刷新、分享、更多选项
2. **左侧边栏**: 会话历史列表，支持删除操作
3. **模型选择区**: 多选模型，查看模型详情
4. **聊天区域**: 消息显示，支持图像展示
5. **输入区域**: 文本输入、图片上传、发送按钮

## 部署说明

### 环境要求
- Go 1.19+
- Node.js 16+
- MySQL 8.0+
- Redis 6.0+

### 配置要求
1. **API密钥配置**:
   ```go
   cfg := &config.Promptt{
       BaseURL:   "https://api.openai.com/v1",
       APIKey:    "your-api-key-here",
       TTSPath:   "/audio/speech",
       PaintPath: "/images/generations",
   }
   ```

2. **数据库配置**:
   - 确保数据库连接正常
   - 运行数据库迁移
   - 初始化默认模型数据

### 后端部署
1. 配置数据库连接
2. 设置API密钥
3. 运行数据库迁移
4. 启动后端服务

### 前端部署
1. 安装依赖: `npm install`
2. 构建项目: `npm run build`
3. 部署到Web服务器

## 集成指南

### 在主应用中集成
```go
// 在现有的main.go中添加
import (
    "github.com/qianlnk/VirtualHumanStudio/backend/routes"
    "github.com/qianlnk/VirtualHumanStudio/backend/models"
    "github.com/qianlnk/VirtualHumanStudio/backend/client/promptt"
)

func main() {
    // 现有的初始化代码...
    
    // 添加聊天功能
    chatRepo := models.NewChatRepository(db)
    prompttClient := promptt.New(prompttConfig)
    routes.SetupChatRoutes(router, chatRepo, prompttClient)
    
    // 启动服务器...
}
```

### 测试API
```bash
# 获取模型列表
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/api/chat/models

# 创建会话
curl -X POST -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"title":"测试会话"}' \
     http://localhost:8080/api/chat/sessions

# 发送消息
curl -X POST -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"session_id":"session_123","message":"画一只猫","models":["DALL-E 3"]}' \
     http://localhost:8080/api/chat/messages
```

## 扩展功能

### 计划中的功能
- [ ] 实时图像生成进度显示
- [ ] 图像编辑和优化功能
- [ ] 批量图像生成
- [ ] 图像风格迁移
- [ ] 社区分享功能
- [ ] 高级参数调整
- [ ] 流式响应支持
- [ ] 多模态输入支持

### 技术优化
- [ ] WebSocket实时通信
- [ ] 图像压缩和优化
- [ ] 缓存机制优化
- [ ] 性能监控和日志
- [ ] 错误重试机制
- [ ] 请求限流控制

## 注意事项

1. **API限制**: 注意各AI服务的API调用限制和费用
2. **错误处理**: 实现适当的错误处理和用户提示
3. **安全性**: 确保API密钥安全存储
4. **性能**: 监控系统性能，优化响应时间
5. **用户体验**: 提供清晰的加载状态和错误提示

## 贡献指南

欢迎提交Issue和Pull Request来改进这个项目！

## 许可证

MIT License 