package ai_v1_user

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"net/http"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
	"strings"
)

// ChatFirstMessage 首次发起会话并聊天
func chatFirstMessage(c *gin.Context) {
	type ChatRequest struct {
		Prompt string `json:"prompt" binding:"required"` // 用户发的首条消息
	}
	type ChatResponse struct {
		Reply       string `json:"reply"`
		SessionCode string `json:"session_code"`
		Title       string `json:"title"`
	}

	var req ChatRequest
	if !utils.ValidationJson(c, &req) {
		return
	}

	userCode := utils.GetUser(c)
	sessionCode := uuid.New().String()

	// 获取 AI 模型配置（唯一）
	var modelCfg model.AIModelConfig
	has, err := config.DB.Get(&modelCfg)
	if err != nil || !has {
		response.Success(c, response.ServerError, errors.New("未找到可用模型配置"))
		return
	}

	// 拼接上下文 Prompt（用于对话回答）
	var fullPrompt strings.Builder
	fullPrompt.WriteString(modelCfg.PromptIntro + "\n")
	fullPrompt.WriteString(fmt.Sprintf("%s：%s\n%s：", modelCfg.UserLabel, req.Prompt, modelCfg.AssistantLabel))

	// 调用 AI 获取回复
	reply, err := SendToDeepSeek("deepseek-r1:1.5b", fullPrompt.String())
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 调用 AI 获取标题
	titlePrompt := fmt.Sprintf("请为下面这段话生成一个简洁、有代表性的中文会话标题：该问题只需要回答结果 不需要任何其他描述\n\n%s", req.Prompt)
	title, err := SendToDeepSeek("deepseek-r1:1.5b", titlePrompt)
	if err != nil {
		title = "未命名会话"
	}
	title = extractTitleFromAIResponse(title)
	// 写入会话表
	session := model.AIChatSession{
		UserCode:     userCode,
		SessionTitle: title,
		LastMessage:  reply,
		BasicModel:   model.BasicModel{Code: sessionCode},
	}
	if _, err := config.DB.Insert(&session); err != nil {
		response.Success(c, response.ServerError, fmt.Errorf("创建会话失败: %v", err))
		return
	}
	//reply = extractTitleFromAIResponse(reply)

	// 写入聊天记录
	SaveChat(userCode, "user", req.Prompt, sessionCode)
	SaveChat(userCode, "assistant", reply, sessionCode)

	response.SuccessWithData(c, response.SuccessCode, ChatResponse{
		Reply:       reply,
		SessionCode: sessionCode,
		Title:       title,
	})

}

// SendToDeepSeek 调用 AI 模型
func SendToDeepSeek(modelName, prompt string) (string, error) {
	url := config.G.AI.Url
	body := map[string]interface{}{
		"model":  modelName,
		"prompt": prompt,
		"stream": false,
	}
	jsonData, _ := json.Marshal(body)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("调用 AI 模型失败: %v", err)
	}
	defer resp.Body.Close()

	resData, _ := io.ReadAll(resp.Body)
	var result struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}
	if err := json.Unmarshal(resData, &result); err != nil {
		return "", fmt.Errorf("解析 AI 响应失败: %v", err)
	}

	return strings.TrimSpace(result.Response), nil
}

// SaveChat 保存聊天记录
func SaveChat(userCode, role, content, sessionCode string) {
	h := model.AIChatHistory{
		UserCode:    userCode,
		SessionCode: sessionCode,
		Role:        role,
		Content:     content,
	}
	h.BeforeInsert()
	_, _ = config.DB.Insert(&h)
}

// 查询用户全部会话
func listAIChatSessions(c *gin.Context) {
	userCode := utils.GetUser(c)
	var sessions []model.AIChatSession
	err := config.DB.Where("user_code = ?", userCode).Desc("updated").Find(&sessions)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}
	response.SuccessWithData(c, response.SuccessCode, sessions)
}

// 删除某个会话
func deleteAIChatSession(c *gin.Context) {
	sessionCode := c.Query("session_code")
	if sessionCode == "" {
		response.Success(c, response.BadRequest, errors.New("缺少 session_id 参数"))
		return
	}
	userCode := utils.GetUser(c)
	_, err := config.DB.Where("code = ? AND user_code = ?", sessionCode, userCode).Delete(&model.AIChatSession{})
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}
	_, _ = config.DB.Where("session_code = ? AND user_code = ?", sessionCode, userCode).Delete(&model.AIChatHistory{})
	response.Success(c, response.SuccessCode)
}

// 修改会话标题
func updateAIChatSessionTitle(c *gin.Context) {
	var req struct {
		SessionCode string `json:"session_code" binding:"required"`
		NewTitle    string `json:"new_title" binding:"required"`
	}

	if !utils.ValidationJson(c, &req) {
		return
	}

	userCode := utils.GetUser(c)
	affected, err := config.DB.
		Where("code = ? AND user_code = ?", req.SessionCode, userCode).
		Update(&model.AIChatSession{SessionTitle: req.NewTitle})
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}
	if affected == 0 {
		response.Success(c, response.BadRequest, errors.New("未找到该会话或无权限"))
		return
	}

	response.Success(c, response.SuccessCode)
}

// ChatInSession 继续某个已有会话的聊天
func chatInSession(c *gin.Context) {
	type ChatRequest struct {
		SessionCode string `json:"session_code" binding:"required"` // 已存在的会话 ID
		Prompt      string `json:"prompt" binding:"required"`       // 用户输入内容
	}

	type ChatResponse struct {
		Reply string `json:"reply"`
	}

	var req ChatRequest
	if !utils.ValidationJson(c, &req) {
		return
	}
	userCode := utils.GetUser(c)

	// 获取模型配置
	var modelCfg model.AIModelConfig
	_, err := config.DB.Get(&modelCfg)
	if err != nil {
		response.Success(c, response.ServerError, errors.New("模型配置读取失败"))
		return
	}

	// 获取历史记录（上下文）
	history := getChatHistory(userCode, req.SessionCode, modelCfg.MaxHistory)

	var fullPrompt strings.Builder
	fullPrompt.WriteString(modelCfg.PromptIntro + "\n")
	for _, h := range history {
		role := modelCfg.UserLabel
		if h.Role == "assistant" {
			role = modelCfg.AssistantLabel
		}
		fullPrompt.WriteString(fmt.Sprintf("%s：%s\n", role, h.Content))
	}
	fullPrompt.WriteString(fmt.Sprintf("%s：%s\n%s：", modelCfg.UserLabel, req.Prompt, modelCfg.AssistantLabel))

	// 调用 AI
	reply, err := SendToDeepSeek("deepseek-r1:1.5b", fullPrompt.String())
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 存储聊天记录
	SaveChat(userCode, "user", req.Prompt, req.SessionCode)
	SaveChat(userCode, "assistant", reply, req.SessionCode)

	// 更新会话最后一句话
	_, _ = config.DB.Where("code = ? AND user_code = ?", req.SessionCode, userCode).
		Update(&model.AIChatSession{LastMessage: reply})

	response.SuccessWithData(c, response.SuccessCode, ChatResponse{
		Reply: reply,
	})
}

// getChatHistory 获取用户某会话的聊天记录（限制条数，按创建时间升序）
func getChatHistory(userCode, sessionCode string, limit int) []model.AIChatHistory {
	var history []model.AIChatHistory
	_ = config.DB.Where("user_code = ? AND session_code = ?", userCode, sessionCode).
		Asc("created").Limit(limit).Find(&history)
	return history
}

// getChatHistoryBySession 查询某会话下的全部聊天记录
func getChatHistoryBySession(c *gin.Context) {
	sessionCode := c.Query("session_code")
	if sessionCode == "" {
		response.Success(c, response.BadRequest, errors.New("缺少 session_code 参数"))
		return
	}

	userCode := utils.GetUser(c)
	var history []model.AIChatHistory
	err := config.DB.Where("user_code = ? AND session_code = ?", userCode, sessionCode).
		Asc("created").Find(&history)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	response.SuccessWithData(c, response.SuccessCode, history)
}

func extractTitleFromAIResponse(resp string) string {
	parts := strings.Split(resp, "</think>")
	if len(parts) > 1 {
		return strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(resp) // fallback，避免空
}
