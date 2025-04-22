package clip

import (
	"time"
)

// 定义统一的文件获取接口
type FileFinder interface {
	// 从剪贴板获取文件路径，无论剪贴板中是文件引用还是文本
	GetFiles(timeout time.Duration) ([]string, error)
}

// 统一的对外接口函数
func GetFiles(timeoutSeconds int) ([]string, error) {
	timeout := time.Duration(timeoutSeconds) * time.Second
	finder := newFileFinder()
	return finder.GetFiles(timeout)
}
