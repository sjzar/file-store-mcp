package mcp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/sjzar/file-store-mcp/internal/storage"
	"github.com/sjzar/file-store-mcp/pkg/clip"
	"github.com/sjzar/file-store-mcp/pkg/version"
)

type Service struct {
	storage *storage.Service
	Server  *server.MCPServer
}

func NewService(storage *storage.Service) *Service {
	s := &Service{
		storage: storage,
		Server:  server.NewMCPServer(Name, version.Version),
	}
	s.Server.AddTool(UploadFilesTool, s.handleUploadFiles)
	s.Server.AddTool(UploadClipboardFilesTool, s.handleUploadClipboardFiles)
	s.Server.AddTool(UploadUrlFilesTool, s.handleUploadUrlFiles)
	return s
}

func (s *Service) handleUploadFiles(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	_paths, ok := request.Params.Arguments["paths"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("path must be a string")
	}
	paths := make([]string, 0, len(_paths))
	for _, _path := range _paths {
		paths = append(paths, _path.(string))
	}

	validatedPaths, err := s.ValidatePaths(paths)
	if err != nil {
		return nil, err
	}

	urls := ""
	for i, path := range validatedPaths {
		_url, err := s.storage.UploadFile(ctx, path)
		if err != nil {
			return nil, err
		}
		urls += fmt.Sprintf("%d: %s\n", i+1, _url)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Upload %d files successfully:\n%s", len(validatedPaths), urls),
			},
		},
	}, nil
}

func (s *Service) handleUploadClipboardFiles(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 从剪贴板获取文件路径，超时时间设为5秒
	paths, err := clip.GetFiles(5)
	if err != nil {
		return nil, fmt.Errorf("failed to get files from clipboard: %w", err)
	}

	if len(paths) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "No files found in clipboard.",
				},
			},
		}, nil
	}

	validatedPaths, err := s.ValidatePaths(paths)
	if err != nil {
		return nil, err
	}

	urls := ""
	for i, path := range validatedPaths {
		_url, err := s.storage.UploadFile(ctx, path)
		if err != nil {
			return nil, err
		}
		urls += fmt.Sprintf("%d: %s\n", i+1, _url)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Upload %d files from clipboard successfully:\n%s", len(validatedPaths), urls),
			},
		},
	}, nil
}

func (s *Service) handleUploadUrlFiles(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	_urls, ok := request.Params.Arguments["urls"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("urls must be an array of strings")
	}

	urls := make([]string, 0, len(_urls))
	for _, _url := range _urls {
		urls = append(urls, _url.(string))
	}

	if len(urls) == 0 {
		return nil, fmt.Errorf("urls cannot be empty")
	}

	resultUrls := ""
	for i, url := range urls {
		// 创建临时文件来保存下载的内容
		tempFile, err := os.CreateTemp("", "download-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp file: %w", err)
		}
		tempPath := tempFile.Name()
		defer os.Remove(tempPath) // 确保临时文件最后被删除

		// 下载文件
		resp, err := http.Get(url)
		if err != nil {
			tempFile.Close()
			return nil, fmt.Errorf("failed to download file from %s: %w", url, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			tempFile.Close()
			return nil, fmt.Errorf("failed to download file from %s: status code %d", url, resp.StatusCode)
		}

		// 将下载的内容写入临时文件
		_, err = io.Copy(tempFile, resp.Body)
		tempFile.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to save downloaded file: %w", err)
		}

		// 上传临时文件
		uploadedUrl, err := s.storage.UploadFile(ctx, tempPath)
		if err != nil {
			return nil, fmt.Errorf("failed to upload file from %s: %w", url, err)
		}

		resultUrls += fmt.Sprintf("%d: %s\n", i+1, uploadedUrl)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Downloaded and uploaded %d files successfully:\n%s", len(urls), resultUrls),
			},
		},
	}, nil
}

func (s *Service) ValidatePaths(paths []string) ([]string, error) {

	validatePaths := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "" {
			return nil, fmt.Errorf("path cannot be empty")
		}

		abs, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("invalid path: %w", err)
		}

		fileInfo, err := os.Stat(abs)
		if err != nil {
			return nil, fmt.Errorf("invalid path: %w", err)
		}

		if fileInfo.IsDir() {
			return nil, fmt.Errorf("path cannot be a directory")
		}
		validatePaths = append(validatePaths, abs)
	}

	return validatePaths, nil
}
