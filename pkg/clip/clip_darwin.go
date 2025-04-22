//go:build darwin && !cocoa
// +build darwin,!cocoa

package clip

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// macOS AppleScript 实现
type darwinAppleScriptFinder struct{}

// 创建 macOS AppleScript 实现的工厂函数
func newFileFinder() FileFinder {
	return &darwinAppleScriptFinder{}
}

// 从剪贴板获取文件路径
func (f *darwinAppleScriptFinder) GetFiles(timeout time.Duration) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// AppleScript 脚本内容
	script := `
try
	-- 获取剪贴板内容
	set clipboardText to do shell script "pbpaste"
	
	-- 定义常用目录列表
	set commonDirs to {¬
		POSIX path of (path to desktop folder), ¬
		POSIX path of (path to documents folder), ¬
		POSIX path of (path to downloads folder), ¬
		POSIX path of (path to pictures folder), ¬
		POSIX path of (path to home folder) & "Movies/", ¬
		POSIX path of (path to home folder) & "Music/" ¬
	}
	
	-- 初始化结果
	set allResults to {}
	
	-- 处理多个文件或单个文件
	if clipboardText contains return then
		-- 分割多个文件名
		set AppleScript's text item delimiters to return
		set fileNames to text items of clipboardText
		set AppleScript's text item delimiters to ""
	else
		-- 单个文件名
		set fileNames to {clipboardText}
	end if
	
	-- 搜索每个文件
	repeat with fileName in fileNames
		-- 初始化找到的标志
		set fileFound to false
		set foundPath to ""
		
		-- 只进行精确匹配搜索
		repeat with dirPath in commonDirs
			set filePath to dirPath & fileName
			set checkCommand to "ls " & quoted form of filePath & " 2>/dev/null || echo ''"
			set checkResult to do shell script checkCommand
			
			if checkResult is not "" then
				set fileFound to true
				set foundPath to filePath
				exit repeat
			end if
		end repeat
		
		-- 如果在常用目录中没找到，尝试使用 mdfind 进行精确文件名匹配
		if not fileFound then
			-- 使用 -name 参数进行精确匹配
			set mdfindCommand to "mdfind -onlyin " & quoted form of (POSIX path of (path to home folder)) & " \"kMDItemDisplayName == '" & fileName & "'\" | head -1 || echo ''"
			set mdfindResult to do shell script mdfindCommand
			
			if mdfindResult is not "" then
				set fileFound to true
				set foundPath to mdfindResult
			end if
		end if
		
		-- 添加结果（只添加找到的文件路径）
		if fileFound then
			set end of allResults to foundPath
		end if
	end repeat
	
	-- 返回结果（一行一个路径）
	set AppleScript's text item delimiters to "%%%DELIMITER%%%"
	return allResults as text
on error
	-- 出错时返回空
	return ""
end try
`

	// 创建命令
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 执行命令
	err := cmd.Run()

	// 检查是否超时
	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("查找文件超时")
	}

	// 检查其他错误
	if err != nil {
		return nil, fmt.Errorf("执行脚本错误: %v, stderr: %s", err, stderr.String())
	}

	// 获取输出并分割为列表
	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return []string{}, nil
	}

	// 使用特殊分隔符分割结果
	paths := strings.Split(output, "%%%DELIMITER%%%")

	// 过滤空字符串
	var filteredPaths []string
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path != "" {
			filteredPaths = append(filteredPaths, path)
		}
	}

	return filteredPaths, nil
}
