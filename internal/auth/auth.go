// Package auth 提供 Google OAuth2 认证功能。
// 使用 PKCE 授权码流程 + 本地回调服务器，
// 凭据和 Token 均存储在 Windows 凭据管理器中（通过 DPAPI 加密）。
package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"time"

	"study/internal/model"

	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	// keyring 服务名
	keyringService = "study-cli"

	// keyring 键名
	keyClientID     = "google-client-id"
	keyClientSecret = "google-client-secret"
	keyToken        = "google-oauth-token"

	// 回调超时
	callbackTimeout = 120 * time.Second
)

// ---------- 凭据管理 ----------

// GetClientIDSecret 从 Windows 凭据管理器读取 Google OAuth 客户端凭据。
// 返回空字符串表示未配置（不视为错误）。
func GetClientIDSecret() (clientID, clientSecret string, err error) {
	clientID, err = keyring.Get(keyringService, keyClientID)
	if err != nil {
		return "", "", nil // 未配置不算错误
	}
	clientSecret, err = keyring.Get(keyringService, keyClientSecret)
	if err != nil {
		return "", "", nil
	}
	return clientID, clientSecret, nil
}

// SaveClientIDSecret 将 Google OAuth 客户端凭据存入 Windows 凭据管理器。
func SaveClientIDSecret(clientID, clientSecret string) error {
	if err := keyring.Set(keyringService, keyClientID, clientID); err != nil {
		return fmt.Errorf("保存 Client ID 失败: %w", err)
	}
	if err := keyring.Set(keyringService, keyClientSecret, clientSecret); err != nil {
		return fmt.Errorf("保存 Client Secret 失败: %w", err)
	}
	return nil
}

// ClearCredentials 清除所有 Google 相关凭据（ClientID/Secret + Token）。
func ClearCredentials() error {
	// 尝试删除所有 key，忽略不存在的错误
	_ = keyring.Delete(keyringService, keyClientID)
	_ = keyring.Delete(keyringService, keyClientSecret)
	_ = keyring.Delete(keyringService, keyToken)
	return nil
}

// ---------- Token 管理 ----------

// GetToken 从 Windows 凭据管理器读取 OAuth2 Token。
func GetToken() (*oauth2.Token, error) {
	data, err := keyring.Get(keyringService, keyToken)
	if err != nil {
		return nil, nil // 未找到不算错误
	}
	var token oauth2.Token
	if err := json.Unmarshal([]byte(data), &token); err != nil {
		return nil, fmt.Errorf("Token 数据损坏，请重新授权: %w", err)
	}
	return &token, nil
}

// SaveToken 将 OAuth2 Token 存入 Windows 凭据管理器。
func SaveToken(token *oauth2.Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("序列化 Token 失败: %w", err)
	}
	return keyring.Set(keyringService, keyToken, string(data))
}

// ClearToken 从 Windows 凭据管理器中删除 OAuth2 Token。
func ClearToken() error {
	return keyring.Delete(keyringService, keyToken)
}

// ---------- 认证状态查询 ----------

// GetAuthInfo 返回当前 Google 服务认证状态。
func GetAuthInfo() model.GoogleAuthInfo {
	info := model.GoogleAuthInfo{}

	clientID, _, err := GetClientIDSecret()
	if err == nil && clientID != "" {
		info.IsConfigured = true
	}

	token, err := GetToken()
	if err == nil && token != nil {
		info.IsAuthorized = token.Valid()
	}

	return info
}

// ---------- 主入口 ----------

// NewHTTPClient 获取已认证的 *http.Client。
//
// 流程：
//  1. 读取凭据管理器中保存的 ClientID/Secret（如未配置则返回 nil, nil）
//  2. 读取已保存的 Token，如果有效则直接返回带自动刷新的 HTTP 客户端
//  3. 如果 Token 不存在或已过期，启动 PKCE 授权流程获取新 Token
func NewHTTPClient(ctx context.Context, scopes []string) (*http.Client, error) {
	clientID, clientSecret, err := GetClientIDSecret()
	if err != nil {
		return nil, fmt.Errorf("读取凭据失败: %w", err)
	}
	if clientID == "" {
		// 未配置 Google 集成，静默返回 nil
		return nil, nil
	}

	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       scopes,
	}

	// 尝试用已保存的 Token
	token, _ := GetToken()
	if token != nil && token.Valid() {
		ts := cfg.TokenSource(ctx, token)
		return oauth2.NewClient(ctx, ts), nil
	}

	// 如果 refresh_token 存在，尝试刷新
	if token != nil && token.RefreshToken != "" {
		ts := cfg.TokenSource(ctx, token)
		newToken, err := ts.Token()
		if err == nil && newToken.Valid() {
			_ = SaveToken(newToken)
			return oauth2.NewClient(ctx, ts), nil
		}
		// 刷新失败，需要重新授权
	}

	// 启动 PKCE 流程
	newToken, err := runPKCEFlow(cfg)
	if err != nil {
		return nil, fmt.Errorf("Google 授权失败: %w", err)
	}

	if err := SaveToken(newToken); err != nil {
		return nil, fmt.Errorf("保存 Token 失败: %w", err)
	}

	ts := cfg.TokenSource(ctx, newToken)
	return oauth2.NewClient(ctx, ts), nil
}

// ---------- PKCE 授权流程 ----------

// runPKCEFlow 启动 PKCE + loopback 本地服务器的 OAuth2 授权流程。
func runPKCEFlow(cfg *oauth2.Config) (*oauth2.Token, error) {
	// 1. 生成 PKCE 参数
	verifier := oauth2.GenerateVerifier()
	state := randomString(32)

	// 2. 在 127.0.0.1 上查找空闲端口
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("无法启动本地回调服务器: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	// 3. 设置重定向 URI
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)
	cfg.RedirectURL = redirectURI

	// 4. 构建授权 URL
	authURL := cfg.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.S256ChallengeOption(verifier),
	)

	// 5. 启动回调服务器
	tokenCh := make(chan *oauth2.Token, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		// 检查错误
		if errStr := query.Get("error"); errStr != "" {
			fmt.Fprintf(w, "<html><body><h1>❌ 授权失败</h1><p>%s</p><p>可以关闭此页面。</p></body></html>", errStr)
			errCh <- fmt.Errorf("授权被拒绝: %s", errStr)
			return
		}

		// 验证 state 防 CSRF
		if query.Get("state") != state {
			http.Error(w, "state 不匹配", http.StatusBadRequest)
			errCh <- fmt.Errorf("安全校验失败: state 不匹配，拒绝请求")
			return
		}

		// 交换 code 获取 token
		code := query.Get("code")
		if code == "" {
			http.Error(w, "缺少授权码", http.StatusBadRequest)
			errCh <- fmt.Errorf("未收到授权码")
			return
		}

		token, err := cfg.Exchange(r.Context(), code, oauth2.VerifierOption(verifier))
		if err != nil {
			fmt.Fprintf(w, "<html><body><h1>❌ Token 交换失败</h1><p>%v</p><p>可以关闭此页面。</p></body></html>", err)
			errCh <- fmt.Errorf("Token 交换失败: %w", err)
			return
		}

		fmt.Fprint(w, "<html><head><meta charset=\"utf-8\"></head><body><h1>✅ 授权成功！</h1><p>可以关闭此页面，回到终端继续使用 study。</p></body></html>")
		tokenCh <- token
	})

	server := &http.Server{Handler: mux}
	go func() {
		_ = server.Serve(listener)
	}()

	// 6. 打开系统浏览器
	fmt.Printf("正在打开浏览器进行 Google 授权...\n")
	fmt.Printf("如果浏览器未自动打开，请手动访问：\n%s\n\n", authURL)
	if err := openBrowser(authURL); err != nil {
		// 浏览器打开失败不是致命错误，用户可手动访问
	}

	// 7. 等待回调
	var token *oauth2.Token
	select {
	case token = <-tokenCh:
		// 成功
	case err := <-errCh:
		_ = server.Shutdown(context.Background())
		return nil, err
	case <-time.After(callbackTimeout):
		_ = server.Shutdown(context.Background())
		return nil, fmt.Errorf("授权超时（%v 秒内未收到浏览器回调）", int(callbackTimeout.Seconds()))
	}

	// 8. 关闭服务器
	_ = server.Shutdown(context.Background())

	return token, nil
}

// ---------- 浏览器打开 ----------

// openBrowser 用系统默认浏览器打开指定 URL（Windows 专用）。
func openBrowser(url string) error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

// ---------- 工具函数 ----------

// randomString 生成指定长度的随机字符串（用于 state 参数防 CSRF）。
func randomString(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}

// ---------- 强制重新授权 ----------

// Reauthorize 强制重新发起 OAuth2 授权（清除旧 Token，启动新 PKCE 流程）。
// 用于 "study google login" 命令。
func Reauthorize(ctx context.Context, scopes []string) error {
	// 清除旧 Token
	_ = ClearToken()

	clientID, clientSecret, err := GetClientIDSecret()
	if err != nil {
		return fmt.Errorf("读取凭据失败: %w", err)
	}
	if clientID == "" {
		return fmt.Errorf("Google 客户端凭据未配置，请先运行 study init")
	}

	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       scopes,
	}

	token, err := runPKCEFlow(cfg)
	if err != nil {
		return err
	}

	// 授权成功后，通知调用方需要重新初始化 Google 服务
	_ = SaveToken(token)
	fmt.Println("✅ Google 授权成功！请重新启动 study 以加载 Google 服务。")
	return nil
}
