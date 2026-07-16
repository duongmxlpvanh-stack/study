package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"study/internal/config"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const (
	// driveRootFolder Google Drive 中的根文件夹名
	driveRootFolder = "study_pdfs"
)

// GoogleDriveService 提供 Google Drive 文件上传和管理功能。
type GoogleDriveService struct {
	cfg        *config.Config
	driveSvc   *drive.Service
	autoUpload bool // 自动上传开关（会话级别，暂不持久化）

	rootFolderID string // 惰性初始化
}

// NewGoogleDriveService 创建 Google Drive 服务实例。
// httpClient 应为已通过 OAuth2 认证的 HTTP 客户端。
func NewGoogleDriveService(cfg *config.Config, httpClient *http.Client) (*GoogleDriveService, error) {
	srv, err := drive.NewService(context.Background(), option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("初始化 Google Drive 服务失败: %w", err)
	}
	return &GoogleDriveService{
		cfg:      cfg,
		driveSvc: srv,
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

	list, err := s.driveSvc.Files.List().Q(query).Fields("files(id, name)").Do()
	if err != nil {
		return "", fmt.Errorf("查询 Drive 文件夹失败: %w", err)
	}
	if len(list.Files) > 0 {
		return list.Files[0].Id, nil
	}

	// 不存在则创建
	folder := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
	}
	if parentID != "" {
		folder.Parents = []string{parentID}
	}

	created, err := s.driveSvc.Files.Create(folder).Fields("id").Do()
	if err != nil {
		return "", fmt.Errorf("创建 Drive 文件夹失败: %w", err)
	}
	return created.Id, nil
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

	// 创建 Drive 文件
	fileName := filepath.Base(localPath)
	driveFile := &drive.File{
		Name:     fileName,
		Parents:  []string{folderID},
		MimeType: detectMimeType(fileName),
	}

	created, err := s.driveSvc.Files.Create(driveFile).Media(file).Fields("id, name, webViewLink").Do()
	if err != nil {
		return "", fmt.Errorf("上传文件到 Google Drive 失败: %w", err)
	}

	return created.Id, nil
}

// ---------- 文件列表 ----------

// ListFiles 列出 Google Drive 中已上传的文件。
// subject 为空时列出所有文件，否则按科目筛选。
func (s *GoogleDriveService) ListFiles(ctx context.Context, subject string) ([]*drive.File, error) {
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
		list, err := s.driveSvc.Files.List().Q(query).
			Fields("files(id, name, size, createdTime, webViewLink)").
			OrderBy("createdTime desc").
			Do()
		if err != nil {
			return nil, fmt.Errorf("列出 Drive 文件失败: %w", err)
		}
		return list.Files, nil
	}

	// 列出所有子文件夹中的所有文件
	query := fmt.Sprintf("'%s' in parents and trashed=false", rootID)
	allFolders, err := s.driveSvc.Files.List().Q(query).
		Fields("files(id, name)").
		Do()
	if err != nil {
		return nil, fmt.Errorf("列出 Drive 文件夹失败: %w", err)
	}

	var allFiles []*drive.File
	for _, folder := range allFolders.Files {
		q := fmt.Sprintf("'%s' in parents and trashed=false", folder.Id)
		list, err := s.driveSvc.Files.List().Q(q).
			Fields("files(id, name, size, createdTime, webViewLink)").
			OrderBy("createdTime desc").
			Do()
		if err != nil {
			continue // 跳过读取失败的文件夹
		}
		// 给每个文件附加父文件夹信息
		for _, f := range list.Files {
			f.Description = folder.Name // 借用 Description 字段存储科目名
		}
		allFiles = append(allFiles, list.Files...)
	}
	return allFiles, nil
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
	folders, err := s.driveSvc.Files.List().Q(query).Fields("files(id, name)").Do()
	if err != nil {
		return nil, fmt.Errorf("查询 Drive 存储信息失败: %w", err)
	}

	info := &StorageInfo{}
	totalFiles := 0
	for _, folder := range folders.Files {
		info.Folders = append(info.Folders, folder.Name)
		q := fmt.Sprintf("'%s' in parents and trashed=false", folder.Id)
		list, err := s.driveSvc.Files.List().Q(q).Fields("files(id)").Do()
		if err == nil {
			totalFiles += len(list.Files)
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
