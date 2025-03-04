package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"regexp"
	"rest/config"
	"rest/response"
)

// 处理 AI 聊天请求
func chatWithAI(c *gin.Context) {
	// 获取用户输入
	var req struct {
		Prompt string `json:"prompt" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Success(c, response.BadRequest, errors.New("缺少 prompt 参数"))
		return
	}

	// 调用 AI 处理
	res, err := SendToAI("deepseek-r1:1.5b", req.Prompt)
	if err != nil {
		response.Success(c, response.ServerError, fmt.Errorf("发送给AI出错: %v", err))
		return
	}

	response.SuccessWithData(c, response.SuccessCode, res)
}

// SendToAI 发送请求到 AI 并返回清理后的客服回复
func SendToAI(model, prompt string) (string, error) {
	type requestBody struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
		Stream bool   `json:"stream"`
	}

	type responseBody struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}

	// **优化 Prompt，告诉 AI 只返回最终客服回答**
	instruct := `你是“阿亮餐厅”的 AI 客服，必须严格遵守以下规则：
1️⃣ **始终保持专业、规范、礼貌的客服语气**。
2️⃣ **使用敬语**（如：您好、请问、感谢您的光临、祝您用餐愉快等）。
3️⃣ **禁止随意闲聊，必须围绕餐厅业务解答问题**。
4️⃣ **回答要简明扼要，不可模棱两可**。
5️⃣ **不得擅自编造餐厅信息，如有疑问请提示用户联系客服**。
6️⃣ **如果用户提出无关或不适当的问题，应礼貌拒绝并引导至正题**。

⚠ **重要：请直接回复最终答案，不要输出任何思考过程，不要包含 <think>...</think> 这样的标签！**
请按照上述规则回答用户的问题：
`

	// 1️⃣ 构造请求体
	reqBody := requestBody{
		Model:  model,
		Prompt: instruct + "\n用户：" + prompt + "\n阿亮餐厅AI客服：",
		Stream: false,
	}

	// 2️⃣ 序列化 JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("❌ JSON 序列化失败: %v", err)
	}

	// 3️⃣ 发送 HTTP 请求
	url := config.G.AI.Url // **你的 AI API 地址**
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("❌ 请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 4️⃣ 读取并解析 AI 响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("❌ 读取 AI 响应失败: %v", err)
	}

	var result responseBody
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("❌ 解析 AI 响应失败: %v", err)
	}

	// 5️⃣ **去除 <think>...</think> 标签**
	cleanedResponse := cleanAIResponse(result.Response)

	// 6️⃣ 只返回 AI 最终的客服回复
	return cleanedResponse, nil
}

// **清理 AI 响应，移除 <think>...</think> 部分**
func cleanAIResponse(response string) string {
	// 使用正则表达式删除 `<think>...</think>` 部分
	re := regexp.MustCompile(`(?s)<think>.*?</think>`)
	cleanedText := re.ReplaceAllString(response, "")

	// **去除首尾空格**
	return cleanedText
}
