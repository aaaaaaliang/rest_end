package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
	"strings"
)

type agentStatus struct {
	IsOnline     bool `json:"is_online"`
	SessionCount int  `json:"session_count"`
}

func createSession(c *gin.Context) {
	type Res struct {
		SessionCode  string `json:"session_code"`
		AgentCode    string `json:"agent_code"`
		AgentName    string `json:"agent_name"`
		CustomerCode string `json:"customer_code"`
	}
	ctx := context.Background()

	// 获取顾客身份
	customerCode := utils.GetUser(c)

	// 1. 检查是否已有进行中的会话
	var existing model.ChatSession
	has, err := config.DB.Where("customer_code = ? AND status = 'active'", customerCode).Get(&existing)
	if err != nil {
		response.Success(c, response.ServerError, fmt.Errorf("查询会话失败: %v", err))
		return
	}
	if has {
		var user model.Users
		_, err = config.DB.Where("code = ?", existing.AgentCode).Get(&user)
		if err != nil {
			response.Success(c, response.ServerError, fmt.Errorf("查找用户: %v", err))
			return
		}
		res := Res{
			SessionCode:  existing.Code,
			AgentCode:    existing.AgentCode,
			AgentName:    user.RealName,
			CustomerCode: existing.CustomerCode,
		}

		// ✅ 已有会话，直接返回，不再走分配逻辑
		response.SuccessWithData(c, response.SuccessCode, res)
		return
	}

	// 2. 无会话，查 Redis 分配在线客服
	pattern := "agent:*"
	keys, err := config.R.Keys(ctx, pattern).Result()
	if err != nil {
		response.Success(c, response.ServerError, fmt.Errorf("读取在线客服失败: %v", err))
		return
	}

	var selectedAgent string
	minSession := int(^uint(0) >> 1) // MaxInt

	for _, key := range keys {
		val, err := config.R.Get(ctx, key).Result()
		if err != nil || val == "" {
			continue
		}
		var status agentStatus
		if json.Unmarshal([]byte(val), &status) != nil || !status.IsOnline {
			continue
		}
		if status.SessionCount < minSession {
			minSession = status.SessionCount
			selectedAgent = strings.TrimPrefix(key, "agent:")
		}
	}

	if selectedAgent == "" {
		// TODO: 此处应进入 AI 或留言模式
		response.Success(c, response.ServerError, errors.New("暂无在线客服，请稍后再试或留言"))
		return
	}

	// 3. 创建新会话
	session := model.ChatSession{
		CustomerCode: customerCode,
		AgentCode:    selectedAgent,
		Status:       "active",
	}
	session.BeforeInsert()
	if _, err := config.DB.Insert(&session); err != nil {
		response.Success(c, response.ServerError, fmt.Errorf("创建会话失败: %v", err))
		return
	}

	var user model.Users
	_, err = config.DB.Where("code = ?", selectedAgent).Get(&user)
	if err != nil {
		response.Success(c, response.ServerError, fmt.Errorf("查找用户: %v", err))
		return
	}

	res := Res{
		SessionCode:  session.Code,
		AgentCode:    selectedAgent,
		AgentName:    user.RealName,
		CustomerCode: customerCode,
	}

	// ✅ 返回创建成功的会话信息
	response.SuccessWithData(c, response.SuccessCode, res)
}

func startSessionByAgent(c *gin.Context) {
	type Req struct {
		CustomerCode string `form:"customer_code" json:"customer_code" binding:"required"`
	}
	type Res struct {
		SessionCode  string `json:"session_code"`
		AgentCode    string `json:"agent_code"`
		CustomerCode string `json:"customer_code"`
		CustomerName string `json:"customer_name"`
	}

	var req Req
	if !utils.ValidationQuery(c, &req) {
		return
	}

	// 获取当前用户身份（客服）
	agentCode := utils.GetUser(c)

	// 确保顾客存在（可选增强）
	has, err := config.DB.Table(model.Users{}).Where("code = ?", req.CustomerCode).Exist()
	if err != nil || !has {
		response.Success(c, response.BadRequest, fmt.Errorf("指定顾客不存在或发生错误: %v", err))
		return
	}

	// 检查是否已有会话
	var existing model.ChatSession
	has, err = config.DB.Where("customer_code = ? AND agent_code = ? AND status = 'active'",
		req.CustomerCode, agentCode).Get(&existing)
	if err != nil {
		response.Success(c, response.ServerError, fmt.Errorf("查询会话失败: %v", err))
		return
	}
	if has {

		var user model.Users
		_, err = config.DB.Where("code = ?", existing.CustomerCode).Get(&user)
		if err != nil {
			response.Success(c, response.ServerError, fmt.Errorf("查找用户: %v", err))
			return
		}
		res := Res{
			SessionCode:  existing.Code,
			AgentCode:    existing.AgentCode,
			CustomerCode: existing.CustomerCode,
			CustomerName: user.RealName,
		}
		// ✅ 已有会话，直接返回，不再走分配逻辑
		response.SuccessWithData(c, response.SuccessCode, res)
		return
	}

	// 创建会话
	session := model.ChatSession{
		CustomerCode: req.CustomerCode,
		AgentCode:    agentCode,
		Status:       "active",
	}
	session.BeforeInsert()
	if _, err := config.DB.Insert(&session); err != nil {
		response.Success(c, response.ServerError, fmt.Errorf("创建会话失败: %v", err))
		return
	}

	var user model.Users
	_, err = config.DB.Where("code = ?", existing.CustomerCode).Get(&user)
	if err != nil {
		response.Success(c, response.ServerError, fmt.Errorf("查找用户: %v", err))
		return
	}
	res := Res{
		SessionCode:  session.Code,
		AgentCode:    agentCode,
		CustomerCode: existing.CustomerCode,
		CustomerName: user.RealName,
	}

	response.SuccessWithData(c, response.SuccessCode, res)
}
