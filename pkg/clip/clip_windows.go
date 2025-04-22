//go:build windows
// +build windows

package clip

import (
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// Windows 实现
type windowsFinder struct{}

// 创建 Windows 实现的工厂函数
func newFileFinder() FileFinder {
	return &windowsFinder{}
}

// Windows API 常量
const (
	CF_UNICODETEXT = 13
	CF_HDROP       = 15
	GMEM_MOVEABLE  = 0x0002
)

// Windows API 函数
var (
	user32                     = syscall.NewLazyDLL("user32.dll")
	kernel32                   = syscall.NewLazyDLL("kernel32.dll")
	shell32                    = syscall.NewLazyDLL("shell32.dll")
	openClipboard              = user32.NewProc("OpenClipboard")
	closeClipboard             = user32.NewProc("CloseClipboard")
	getClipboardData           = user32.NewProc("GetClipboardData")
	isClipboardFormatAvailable = user32.NewProc("IsClipboardFormatAvailable")
	globalLock                 = kernel32.NewProc("GlobalLock")
	globalUnlock               = kernel32.NewProc("GlobalUnlock")
	dragQueryFileW             = shell32.NewProc("DragQueryFileW")
)

// 从剪贴板获取文件路径
func (f *windowsFinder) GetFiles(timeout time.Duration) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resultChan := make(chan []string, 1)
	errChan := make(chan error, 1)

	go func() {
		// 首先尝试获取剪贴板中的文件引用
		ret, _, _ := openClipboard.Call(0)
		if ret == 0 {
			errChan <- fmt.Errorf("打开剪贴板失败")
			return
		}
		defer closeClipboard.Call()

		// 检查是否有文件格式数据
		isFormatAvailable, _, _ := isClipboardFormatAvailable.Call(uintptr(CF_HDROP))
		if isFormatAvailable != 0 {
			h, _, _ := getClipboardData.Call(uintptr(CF_HDROP))
			if h == 0 {
				errChan <- fmt.Errorf("获取剪贴板数据失败")
				return
			}

			ptr, _, _ := globalLock.Call(h)
			if ptr == 0 {
				errChan <- fmt.Errorf("锁定剪贴板内存失败")
				return
			}
			defer globalUnlock.Call(h)

			// 获取文件数量
			fileCount, _, _ := dragQueryFileW.Call(ptr, 0xFFFFFFFF, 0, 0)

			var paths []string
			for i := uint(0); i < uint(fileCount); i++ {
				// 获取所需缓冲区大小
				bufSize, _, _ := dragQueryFileW.Call(ptr, uintptr(i), 0, 0)
				bufSize += 1 // 加上 null 终止符

				// 创建缓冲区
				buf := make([]uint16, bufSize)
				dragQueryFileW.Call(ptr, uintptr(i), uintptr(unsafe.Pointer(&buf[0])), uintptr(bufSize))

				// 转换为 Go 字符串
				path := syscall.UTF16ToString(buf)
				paths = append(paths, path)
			}

			resultChan <- paths
			return
		}

		// 如果没有文件引用，尝试获取剪贴板文本
		isFormatAvailable, _, _ = isClipboardFormatAvailable.Call(uintptr(CF_UNICODETEXT))
		if isFormatAvailable == 0 {
			// 剪贴板中没有文本，返回空结果
			resultChan <- []string{}
			return
		}

		h, _, _ := getClipboardData.Call(uintptr(CF_UNICODETEXT))
		if h == 0 {
			errChan <- fmt.Errorf("获取剪贴板文本失败")
			return
		}

		ptr, _, _ := globalLock.Call(h)
		if ptr == 0 {
			errChan <- fmt.Errorf("锁定剪贴板内存失败")
			return
		}
		defer globalUnlock.Call(h)

		// 获取剪贴板文本
		clipboardText := syscall.UTF16ToString((*[1 << 20]uint16)(unsafe.Pointer(ptr))[:])
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
		// Windows 路径格式: C:\path\to\file 或 \\server\share\path
		if (len(line) >= 3 && line[1] == ':' && (line[2] == '\\' || line[2] == '/')) ||
			(strings.HasPrefix(line, "\\\\")) {
			// 规范化路径分隔符
			path := strings.ReplaceAll(line, "/", "\\")
			paths = append(paths, path)
		}
	}

	return paths
}
