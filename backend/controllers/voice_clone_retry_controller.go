package controllers

import (
	"net/http"

	"VirtualHumanStudio/backend/db"
	"VirtualHumanStudio/backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RetryVoiceClone 重试音色克隆任务
func RetryVoiceClone(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 获取任务ID
	id := c.Param("id")

	// 查询任务
	var voiceClone models.VoiceClone
	result := db.DB.First(&voiceClone, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "音色克隆任务不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + result.Error.Error()})
		}
		return
	}

	// 检查权限
	if voiceClone.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作此任务"})
		return
	}

	// 检查任务状态
	if voiceClone.Status != "failed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只能重试失败的任务"})
		return
	}

	// 重置任务状态
	updates := map[string]interface{}{
		"status":    "pending",
		"error_msg": "",
	}
	result = db.DB.Model(&voiceClone).Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新任务状态失败: " + result.Error.Error()})
		return
	}

	// 重新提交任务
	go func() {
		// 更新状态为处理中
		db.DB.Model(&voiceClone).Update("status", "processing")

		// 调用音色克隆API
		apiReq := APIVoiceCloneRequest{
			ModelName:   voiceClone.ModelName,
			PromptFile:  voiceClone.PromptFile,
			PromptText:  voiceClone.PromptText,
			SpeakerName: voiceClone.SpeakerName,
		}

		// 重用现有的API调用逻辑
		handleVoiceCloneAPI(apiReq, &voiceClone)
	}()

	c.JSON(http.StatusOK, gin.H{"message": "任务已重新提交"})
}
