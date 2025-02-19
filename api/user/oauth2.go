package user

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"log"
	"net/http"

	"rest/config"
	"rest/model"
	"rest/response"
)

// **生成随机 state**
func generateState() string {
	stateBytes := make([]byte, 32)
	_, err := rand.Read(stateBytes)
	if err != nil {
		log.Fatal("Failed to generate random state:", err)
	}
	return base64.URLEncoding.EncodeToString(stateBytes)
}

// **GitHub 登录请求**
func githubLogin(c *gin.Context) {
	oauth2Config := oauth2.Config{
		ClientID:     config.G.Oauth2.ClientID,
		ClientSecret: config.G.Oauth2.ClientSecret,
		RedirectURL:  config.G.Oauth2.RedirectURI,
		Scopes:       []string{config.G.Oauth2.Scope},
		Endpoint:     github.Endpoint,
	}
	log.Println("我被调用了")

	oauth2StateString := generateState() // 生成随机 state
	c.SetCookie("oauth2_state", oauth2StateString, 3600, "/", "", false, true)

	// 生成 GitHub OAuth2 授权 URL
	url := oauth2Config.AuthCodeURL(oauth2StateString, oauth2.AccessTypeOffline)
	// 重定向到 GitHub 授权页面
	c.Redirect(http.StatusFound, url)
}

// **GitHub 回调处理**
func githubCallback(c *gin.Context) {
	// 获取 GitHub 返回的授权码和 state
	code := c.Query("code")
	state := c.Query("state")

	// 从 cookie 中获取存储的 state
	cookieState, err := c.Cookie("oauth2_state")
	if err != nil {
		response.Success(c, response.Unauthorized, errors.New("state cookie not found"))
		return
	}

	if state != cookieState {
		response.Success(c, response.Unauthorized, errors.New("invalid state"))
		return
	}

	oauth2Config := oauth2.Config{
		ClientID:     config.G.Oauth2.ClientID,
		ClientSecret: config.G.Oauth2.ClientSecret,
		RedirectURL:  config.G.Oauth2.RedirectURI,
		Scopes:       []string{config.G.Oauth2.Scope},
		Endpoint:     github.Endpoint,
	}

	// **使用授权码获取 Access Token**
	token, err := oauth2Config.Exchange(c, code)
	if err != nil {
		response.Success(c, response.Unauthorized, fmt.Errorf("failed to exchange token %v", err))
		return
	}

	// **存储 Token 到 Cookie**
	c.SetCookie("oauth2_token", token.AccessToken, 3600, "/", "", false, true)

	// **使用 Access Token 获取用户信息**
	client := oauth2Config.Client(c, token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		response.Success(c, response.Unauthorized, errors.New("failed to get user info"))
		return
	}
	defer resp.Body.Close()

	// **解析 GitHub 用户信息**
	var user map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		response.Success(c, response.Unauthorized, errors.New("failed to decode user info"))
		return
	}

	// **获取 GitHub 用户名**
	username, ok := user["login"].(string)
	if !ok {
		response.Success(c, response.BadRequest, errors.New("invalid GitHub username"))
		return
	}

	// **检查用户名是否存在**
	var userInfo model.Users
	exist, err := config.DB.Where("username = ?", username).Get(&userInfo)
	if err != nil {
		response.Success(c, response.ServerError, fmt.Errorf("数据库查询失败: %v", err))
		return
	}

	if !exist {
		// **获取 GitHub 邮箱（可能为空）**
		email, _ := user["email"].(string)

		// **存储用户数据到数据库**
		newUser := model.Users{
			Username: username,
			Email:    email,
			Nickname: user["name"].(string),
		}

		affectRow, err := config.DB.Insert(&newUser)
		if err != nil || affectRow != 1 {
			response.Success(c, response.ServerError, fmt.Errorf("GitHub 用户存储失败: %v", err))
			return
		}

		// 然后再查用户code
		var u model.Users
		_, _ = config.DB.Where("username = ?", username).Get(&u)
		userInfo.Code = u.Code
	}

	url := config.G.App.Front
	var userRole []model.UserRole
	if err = config.DB.Where("user_code = ?", userInfo.Code).Find(&userRole); err != nil {
		response.Success(c, response.ServerError, fmt.Errorf("查询角色: %v", err))
		return
	}
	if len(userRole) > 0 {
		url += "/admin"
	}

	// 生成 JWT Token
	jwt, err := config.GenerateJWT(userInfo.Code)
	if err != nil {
		response.Success(c, response.ServerError, errors.New("生成 Token 失败"))
		return
	}
	// 设置 Cookie
	c.SetCookie("access_token", jwt, 3600, "/", "", false, true)

	c.Redirect(http.StatusFound, url)
}
