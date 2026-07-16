package model

// GoogleAuthInfo 表示当前 Google 服务的认证状态
type GoogleAuthInfo struct {
	IsConfigured bool     // ClientID/ClientSecret 已保存在凭据管理器中
	IsAuthorized bool     // Token 存在且有效
	Scopes       []string // 已授权的权限范围
	Email        string   // Google 账号邮箱（从 token 信息获取）
}
