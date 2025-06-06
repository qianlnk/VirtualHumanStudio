# 会员中间件使用说明

## 概述

会员中间件用于检查用户的会员权限和使用限制，包括通用的使用次数限制和特定功能的使用限制。

## 中间件类型

### 1. 通用中间件

- `MembershipCheck()`: 检查用户的通用使用次数限制
- `MembershipAccess(minLevel string)`: 检查用户的会员级别是否满足要求

### 2. 特定功能中间件

- `VoiceCloneCheck()`: 检查音色克隆功能的使用限制
- `TTSCheck()`: 检查语音合成功能的使用限制
- `ASRCheck()`: 检查语音识别功能的使用限制
- `DigitalHumanCheck()`: 检查数字人合成功能的使用限制
- `ImageProcessCheck()`: 检查图片处理功能的使用限制

### 3. 使用量记录中间件

- `IncrementFeatureUsage(usageValue int)`: 增加特定功能的使用量，需要与特定功能中间件配合使用

## 使用示例

```go
// 路由组使用通用会员检查
authorized := r.Group("/api")
authorized.Use(middleware.MembershipCheck())

// 高级功能路由组需要付费会员
premium := authorized.Group("/premium")
premium.Use(middleware.MembershipAccess("monthly"))

// 音色克隆功能路由
voiceClone := authorized.Group("/voice-clone")
voiceClone.Use(middleware.VoiceCloneCheck())
// 记录使用量，参数为0表示只记录次数不记录数值
voiceClone.POST("/create", middleware.IncrementFeatureUsage(0), controllers.CreateVoiceClone)

// 语音合成功能路由，记录字数
tts := authorized.Group("/tts")
tts.Use(middleware.TTSCheck())
// 在处理函数中获取文本长度并记录
tts.POST("/synthesize", func(c *gin.Context) {
    // 获取请求参数
    var req struct {
        Text string `json:"text"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
        return
    }
    
    // 设置使用值（文本长度）
    c.Set("usage_value", len(req.Text))
    c.Next()
}, middleware.IncrementFeatureUsage(0), controllers.SynthesizeSpeech)
```

## 错误响应

当用户超出使用限制时，中间件会返回以下错误响应：

```json
{
  "error": "您今日的音色克隆次数已达上限，请升级会员或明日再试",
  "code": "feature_usage_limit_exceeded"
}
```

当用户会员级别不满足要求时，中间件会返回以下错误响应：

```json
{
  "error": "此功能需要会员才能使用",
  "code": "membership_required"
}
```

当用户会员已过期时，中间件会返回以下错误响应：

```json
{
  "error": "您的会员已过期，请续费",
  "code": "membership_expired"
}
```