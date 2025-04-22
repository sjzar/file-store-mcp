//go:build linux
// +build linux

package clip

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Linux 实现
type linuxFinder struct{}

// 创建 Linux 实现的工厂函数
func newFileFinder() FileFinder {
	return &linuxFinder{}
}

// 从剪贴板获取文件路径
func (f *linuxFinder) GetFiles(timeout time.Duration) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resultChan := make(chan []string, 1)
	errChan := make(chan error, 1)

	go func() {
		// 首先尝试从剪贴板获取文件 URI
		var cmd *exec.Cmd
		var output []byte
		var err error

		// 尝试使用 xclip 获取 URI 列表 (x11)
		cmd = exec.Command("xclip", "-selection", "clipboard", "-t", "text/uri-list", "-o")
		output, err = cmd.Output()

		// 如果 xclip 失败，尝试使用 wl-paste (Wayland)
		if err != nil {
			cmd = exec.Command("wl-paste", "-t", "text/uri-list")
			output, err = cmd.Output()
		}

		// 如果成功获取 URI 列表
		if err == nil && len(output) > 0 {
			uriList := strings.TrimSpace(string(output))
			if uriList != "" {
				var paths []string
				for _, uri := range strings.Split(uriList, "\n") {
					uri = strings.TrimSpace(uri)
					if uri == "" || strings.HasPrefix(uri, "#") {
						continue
					}

					// 将 file:// URI 转换为路径
					if strings.HasPrefix(uri, "file://") {
						path := strings.TrimPrefix(uri, "file://")
						// 处理 URL 编码
						path = strings.ReplaceAll(path, "%20", " ")
						// 处理其他常见的 URL 编码
						path = strings.ReplaceAll(path, "%25", "%")
						path = strings.ReplaceAll(path, "%23", "#")
						path = strings.ReplaceAll(path, "%26", "&")
						path = strings.ReplaceAll(path, "%2B", "+")
						path = strings.ReplaceAll(path, "%2C", ",")
						path = strings.ReplaceAll(path, "%3A", ":")
						path = strings.ReplaceAll(path, "%3B", ";")
						path = strings.ReplaceAll(path, "%3D", "=")
						path = strings.ReplaceAll(path, "%3F", "?")
						path = strings.ReplaceAll(path, "%40", "@")

						paths = append(paths, path)
					}
				}

				if len(paths) > 0 {
					resultChan <- paths
					return
				}
			}
		}

		// 如果没有文件引用，尝试获取剪贴板文本
		// 尝试使用 xclip 获取文本
		cmd = exec.Command("xclip", "-selection", "clipboard", "-o")
		output, err = cmd.Output()

		// 如果 xclip 失败，尝试使用 xsel
		if err != nil {
			cmd = exec.Command("xsel", "--clipboard", "--output")
			output, err = cmd.Output()
		}

		// 如果 xsel 失败，尝试使用 wl-paste
		if err != nil {
			cmd = exec.Command("wl-paste")
			output, err = cmd.Output()
		}

		// 如果获取剪贴板文本失败，返回空结果
		if err != nil {
			resultChan <- []string{}
			return
		}

		clipboardText := strings.TrimSpace(string(output))
		if clipboardText == "" {
			resultChan <- []string{}
			return
		}

		// 解析剪贴板文本为可能的文件路径
		paths := parseFilePaths(clipboardText)

		// 验证这些路径是否存在
		var validPaths []string
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				validPaths = append(validPaths, path)
			}
		}

		resultChan <- validPaths
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("获取文件路径超时")
	case result := <-resultChan:
		return result, nil
	case err := <-errChan:
		return nil, err
	}
}

// 解析剪贴板文本为可能的文件路径
func parseFilePaths(text string) []string {
	if text == "" {
		return []string{}
	}

	// 分割文本为行
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")

	// 过滤空行并检查是否是有效路径格式
	var paths []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 检查是否是文件路径格式
		// Linux 路径格式: /path/to/file 或 ~/path/to/file
		if strings.HasPrefix(line, "/") || strings.HasPrefix(line, "~/") {
			// 处理 ~ 符号
			if strings.HasPrefix(line, "~/") {
				homeDir, err := os.UserHomeDir()
				if err == nil {
					line = homeDir + line[1:]
				}
			}
			paths = append(paths, line)
		}
	}

	return paths
}
