package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"study/internal/config"
)

const (
	// driveRootFolder Google Drive 中的根文件夹名
	driveRootFolder = "study_pdfs"

	// Google Drive REST API 端点
	driveAPIURL  = "https://www.googleapis.com/drive/v3/files"
	driveUploadURL = "https://www.googleapis.com/upload/drive/v3/files"
)

// ---------- 自定义类型 ----------

// DriveFile Google Drive 文件信息。
// 字段名与旧版 Drive API 客户端兼容，不改变 CLI 层代码。
type DriveFile struct {
	ID          string
	Name        string
	Size        int64
	CreatedTime string
	WebViewLink string
	Description string // 借用存储科目名（兼容现有 CLI 代码）
}

// jsonDriveFile Google Drive API v3 返回的原始 JSON 结构。
// Size 字段在 API 响应中为字符串（如 "1024"），需要转为 int64。
type jsonDriveFile struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Size        string `json:"size"`
	CreatedTime string `json:"createdTime"`
	WebViewLink string `json:"webViewLink"`
}

// jsonDriveFilesList Drive API 列表查询的 JSON 响应。
type jsonDriveFilesList struct {
	Files []jsonDriveFile `json:"files"`
}

// toDriveFile 将 API 响应的 JSON 结构转为公开的 DriveFile 类型。
func (f *jsonDriveFile) toDriveFile() *DriveFile {
	size, _ := strconv.ParseInt(f.Size, 10, 64)
	return &DriveFile{
		ID:          f.ID,
		Name:        f.Name,
		Size:        size,
		CreatedTime: f.CreatedTime,
		WebViewLink: f.WebViewLink,
	}
}

// ---------- 服务结构 ----------

// GoogleDriveService 提供 Google Drive 文件上传和管理功能。
type GoogleDriveService struct {
	cfg        *config.Config
	httpClient *http.Client
	autoUpload bool // 自动上传开关（会话级别，暂不持久化）

	rootFolderID string // 惰性初始化
}

// NewGoogleDriveService 创建 Google Drive 服务实例。
// httpClient 应为已通过 OAuth2 认证的 HTTP 客户端（自动附带 Bearer token）。
func NewGoogleDriveService(cfg *config.Config, httpClient *http.Client) (*GoogleDriveService, error) {
	return &GoogleDriveService{
		cfg:        cfg,
		httpClient: httpClient,
	}, nil
}

// ---------- 自动上传开关 ----------

// ToggleAutoUpload 开关自动上传功能。
func (s *GoogleDriveService) ToggleAutoUpload(enabled bool) {
	s.autoUpload = enabled
}

// IsAutoUpload 返回当前自动上传开关状态。
func (s *GoogleDriveService) IsAutoUpload() bool {
	return s.autoUpload
}

// ---------- 文件夹管理 ----------

// ensureRootFolder 确保 Drive 中存在 study_pdfs 根文件夹。
func (s *GoogleDriveService) ensureRootFolder(ctx context.Context) (string, error) {
	if s.rootFolderID != "" {
		return s.rootFolderID, nil
	}
	id, err := s.findOrCreateFolder(ctx, driveRootFolder, "")
	if err != nil {
		return "", err
	}
	s.rootFolderID = id
	return id, nil
}

// findOrCreateFolder 在 Drive 中查找或创建指定名称的文件夹。
// parentID 为空字符串时在根目录创建。
func (s *GoogleDriveService) findOrCreateFolder(ctx context.Context, name, parentID string) (string, error) {
	// 先查找
	query := fmt.Sprintf("name='%s' and mimeType='application/vnd.google-apps.folder' and trashed=false", escapeQuery(name))
	if parentID != "" {
		query += fmt.Sprintf(" and '%s' in parents", parentID)
	} else {
		query += " and 'root' in parents"
	}

	reqURL := fmt.Sprintf("%s?q=%s&fields=files(id,name)", driveAPIURL, url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("构建 Drive 查询请求失败: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("查询 Drive 文件夹失败: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取 Drive 查询响应失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("查询 Drive 文件夹失败 (%d): %s", resp.StatusCode, string(respBytes))
	}

	var listResp jsonDriveFilesList
	if err := json.Unmarshal(respBytes, &listResp); err != nil {
		return "", fmt.Errorf("解析 Drive 响应失败: %w", err)
	}
	if len(listResp.Files) > 0 {
		return listResp.Files[0].ID, nil
	}

	// 不存在则创建
	folderData := map[string]interface{}{
		"name":     name,
		"mimeType": "application/vnd.google-apps.folder",
	}
	if parentID != "" {
		folderData["parents"] = []string{parentID}
	}

	createBody, err := json.Marshal(folderData)
	if err != nil {
		return "", fmt.Errorf("序列化文件夹 JSON 失败: %w", err)
	}

	req, err = http.NewRequestWithContext(ctx, http.MethodPost, driveAPIURL+"?fields=id", bytes.NewReader(createBody))
	if err != nil {
		return "", fmt.Errorf("构建创建文件夹请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("创建 Drive 文件夹失败: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取创建文件夹响应失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("创建 Drive 文件夹失败 (%d): %s", resp.StatusCode, string(respBytes))
	}

	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respBytes, &created); err != nil {
		return "", fmt.Errorf("解析创建文件夹响应失败: %w", err)
	}
	return created.ID, nil
}

// ---------- 文件上传 ----------

// UploadFile 上传本地文件到 Google Drive 对应科目的文件夹中。
// 文件存放在 study_pdfs/<subject>/<filename>。
// subject 为空时使用"未分类"。
func (s *GoogleDriveService) UploadFile(ctx context.Context, localPath, subject string) (string, error) {
	// 验证文件存在
	info, err := os.Stat(localPath)
	if err != nil {
		return "", fmt.Errorf("文件不存在: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("不支持上传文件夹: %s", localPath)
	}

	// 确保根文件夹存在
	rootID, err := s.ensureRootFolder(ctx)
	if err != nil {
		return "", err
	}

	// 确保科目文件夹存在
	if subject == "" {
		subject = "未分类"
	}
	folderID, err := s.findOrCreateFolder(ctx, subject, rootID)
	if err != nil {
		return "", err
	}

	// 打开文件
	file, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 构建 multipart/related 请求体
	fileName := filepath.Base(localPath)
	mimeType := detectMimeType(fileName)

	metadata := map[string]interface{}{
		"name":     fileName,
		"parents":  []string{folderID},
		"mimeType": mimeType,
	}
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return "", fmt.Errorf("序列化文件元数据失败: %w", err)
	}

	var bodyBuf bytes.Buffer
	boundary := "study_upload_boundary"

	// Part 1: 文件元数据 JSON
	bodyBuf.WriteString("--" + boundary + "\r\n")
	bodyBuf.WriteString("Content-Type: application/json; charset=UTF-8\r\n")
	bodyBuf.WriteString("\r\n")
	bodyBuf.Write(metadataBytes)
	bodyBuf.WriteString("\r\n")

	// Part 2: 文件内容
	bodyBuf.WriteString("--" + boundary + "\r\n")
	bodyBuf.WriteString("Content-Type: " + mimeType + "\r\n")
	bodyBuf.WriteString("\r\n")
	if _, err := io.Copy(&bodyBuf, file); err != nil {
		return "", fmt.Errorf("读取文件内容失败: %w", err)
	}
	bodyBuf.WriteString("\r\n")

	// 结束边界
	bodyBuf.WriteString("--" + boundary + "--\r\n")

	// 发送请求
	reqURL := driveUploadURL + "?uploadType=multipart"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, &bodyBuf)
	if err != nil {
		return "", fmt.Errorf("构建上传请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "multipart/related; boundary="+boundary)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("上传文件到 Google Drive 失败: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取上传响应失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("上传文件到 Google Drive 失败 (%d): %s", resp.StatusCode, string(respBytes))
	}

	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respBytes, &created); err != nil {
		return "", fmt.Errorf("解析上传响应失败: %w", err)
	}

	return created.ID, nil
}

// ---------- 文件列表 ----------

// ListFiles 列出 Google Drive 中已上传的文件。
// subject 为空时列出所有文件，否则按科目筛选。
func (s *GoogleDriveService) ListFiles(ctx context.Context, subject string) ([]*DriveFile, error) {
	rootID, err := s.ensureRootFolder(ctx)
	if err != nil {
		return nil, err
	}

	if subject != "" {
		// 找到科目文件夹
		folderID, err := s.findOrCreateFolder(ctx, subject, rootID)
		if err != nil {
			return nil, err
		}
		query := fmt.Sprintf("'%s' in parents and trashed=false", folderID)
		reqURL := fmt.Sprintf("%s?q=%s&fields=files(id,name,size,createdTime,webViewLink)&orderBy=createdTime desc",
			driveAPIURL, url.QueryEscape(query))

		files, err := s.listFilesByURL(ctx, reqURL)
		if err != nil {
			return nil, fmt.Errorf("列出 Drive 文件失败: %w", err)
		}
		return files, nil
	}

	// 列出所有子文件夹中的所有文件
	query := fmt.Sprintf("'%s' in parents and trashed=false", rootID)
	reqURL := fmt.Sprintf("%s?q=%s&fields=files(id,name)", driveAPIURL, url.QueryEscape(query))

	allFolders, err := s.listFilesByURL(ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("列出 Drive 文件夹失败: %w", err)
	}

	var allFiles []*DriveFile
	for _, folder := range allFolders {
		q := fmt.Sprintf("'%s' in parents and trashed=false", folder.ID)
		u := fmt.Sprintf("%s?q=%s&fields=files(id,name,size,createdTime,webViewLink)&orderBy=createdTime desc",
			driveAPIURL, url.QueryEscape(q))
		list, err := s.listFilesByURL(ctx, u)
		if err != nil {
			continue // 跳过读取失败的文件夹
		}
		// 给每个文件附加父文件夹信息
		for _, f := range list {
			f.Description = folder.Name // 借用 Description 字段存储科目名
		}
		allFiles = append(allFiles, list...)
	}
	return allFiles, nil
}

// listFilesByURL 发送 GET 请求获取文件列表，返回 []*DriveFile。
func (s *GoogleDriveService) listFilesByURL(ctx context.Context, reqURL string) ([]*DriveFile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 错误 (%d): %s", resp.StatusCode, string(respBytes))
	}

	var listResp jsonDriveFilesList
	if err := json.Unmarshal(respBytes, &listResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	var result []*DriveFile
	for i := range listResp.Files {
		result = append(result, listResp.Files[i].toDriveFile())
	}
	return result, nil
}

// ---------- 状态查询 ----------

// StorageInfo 返回 Drive 存储使用情况。
type StorageInfo struct {
	TotalFiles int
	Folders    []string // 科目文件夹列表
}

// GetStorageInfo 获取 Drive 存储使用信息。
func (s *GoogleDriveService) GetStorageInfo(ctx context.Context) (*StorageInfo, error) {
	rootID, err := s.ensureRootFolder(ctx)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("'%s' in parents and mimeType='application/vnd.google-apps.folder' and trashed=false", rootID)
	reqURL := fmt.Sprintf("%s?q=%s&fields=files(id,name)", driveAPIURL, url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("构建查询请求失败: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("查询 Drive 存储信息失败: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("查询 Drive 存储信息失败 (%d): %s", resp.StatusCode, string(respBytes))
	}

	var foldersResp jsonDriveFilesList
	if err := json.Unmarshal(respBytes, &foldersResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	info := &StorageInfo{}
	totalFiles := 0
	for _, folder := range foldersResp.Files {
		info.Folders = append(info.Folders, folder.Name)
		q := fmt.Sprintf("'%s' in parents and trashed=false", folder.ID)
		u := fmt.Sprintf("%s?q=%s&fields=files(id)", driveAPIURL, url.QueryEscape(q))
		listFiles, err := s.listFilesByURL(ctx, u)
		if err == nil {
			totalFiles += len(listFiles)
		}
	}
	info.TotalFiles = totalFiles
	return info, nil
}

// ---------- 辅助函数 ----------

// detectMimeType 根据文件扩展名返回 MIME 类型。
func detectMimeType(fileName string) string {
	ext := filepath.Ext(fileName)
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".ppt":
		return "application/vnd.ms-powerpoint"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".txt":
		return "text/plain"
	case ".md":
		return "text/markdown"
	default:
		return "application/octet-stream"
	}
}

// escapeQuery 转义 Drive 查询中的特殊字符（单引号）。
func escapeQuery(s string) string {
	result := ""
	for _, ch := range s {
		if ch == '\'' {
			result += "\\'"
		} else if ch == '\\' {
			result += "\\\\"
		} else {
			result += string(ch)
		}
	}
	return result
}

// Close 关闭服务（当前无资源需释放，预留接口）。
func (s *GoogleDriveService) Close() error {
	return nil
}
