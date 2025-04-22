//go:build darwin && cocoa
// +build darwin,cocoa

package clip

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

// 从剪贴板获取文件路径
char** getClipboardFilePaths(int* count) {
    NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
    NSArray *classes = @[[NSURL class]];
    NSDictionary *options = @{NSPasteboardURLReadingFileURLsOnlyKey: @YES};

    NSArray *urls = [pasteboard readObjectsForClasses:classes options:options];
    if (urls == nil || [urls count] == 0) {
        *count = 0;
        return NULL;
    }

    *count = (int)[urls count];
    char** result = (char**)malloc(sizeof(char*) * [urls count]);

    for (int i = 0; i < [urls count]; i++) {
        NSURL *url = urls[i];
        NSString *path = [url path];
        const char *cPath = [path UTF8String];
        result[i] = strdup(cPath);
    }

    return result;
}

// 获取剪贴板文本内容
char* getClipboardText() {
    NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
    NSString *string = [pasteboard stringForType:NSPasteboardTypeString];
    if (string == nil) {
        return NULL;
    }

    const char *cString = [string UTF8String];
    return strdup(cString);
}

// 根据文件名查找文件路径
char** findFilesByNames(const char** fileNames, int fileCount, int* resultCount) {
    NSMutableArray *results = [NSMutableArray array];

    // 获取常用目录
    NSArray *commonDirs = @[
        NSSearchPathForDirectoriesInDomains(NSDesktopDirectory, NSUserDomainMask, YES)[0],
        NSSearchPathForDirectoriesInDomains(NSDocumentDirectory, NSUserDomainMask, YES)[0],
        NSSearchPathForDirectoriesInDomains(NSDownloadsDirectory, NSUserDomainMask, YES)[0],
        NSSearchPathForDirectoriesInDomains(NSPicturesDirectory, NSUserDomainMask, YES)[0],
        NSSearchPathForDirectoriesInDomains(NSMoviesDirectory, NSUserDomainMask, YES)[0],
        NSSearchPathForDirectoriesInDomains(NSMusicDirectory, NSUserDomainMask, YES)[0]
    ];

    NSFileManager *fileManager = [NSFileManager defaultManager];

    for (int i = 0; i < fileCount; i++) {
        NSString *fileName = [NSString stringWithUTF8String:fileNames[i]];
        BOOL fileFound = NO;

        // 在常用目录中搜索
        for (NSString *dirPath in commonDirs) {
            NSString *filePath = [dirPath stringByAppendingPathComponent:fileName];
            if ([fileManager fileExistsAtPath:filePath]) {
                [results addObject:filePath];
                fileFound = YES;
                break;
            }
        }

        // 如果没找到，使用Spotlight搜索
        if (!fileFound) {
            NSMetadataQuery *query = [[NSMetadataQuery alloc] init];
            [query setPredicate:[NSPredicate predicateWithFormat:@"kMDItemDisplayName == %@", fileName]];
            [query setSearchScopes:@[NSMetadataQueryLocalComputerScope]];

            // 同步执行查询
            [query startQuery];
            [[NSRunLoop currentRunLoop] runUntilDate:[NSDate dateWithTimeIntervalSinceNow:1.0]];
            [query stopQuery];

            if ([query resultCount] > 0) {
                NSMetadataItem *item = [query resultAtIndex:0];
                NSString *path = [item valueForAttribute:NSMetadataItemPathKey];
                [results addObject:path];
            }
        }
    }

    // 转换结果为C字符串数组
    *resultCount = (int)[results count];
    if (*resultCount == 0) {
        return NULL;
    }

    char **resultArray = (char**)malloc(sizeof(char*) * (*resultCount));
    for (int i = 0; i < *resultCount; i++) {
        const char *utf8Path = [[results objectAtIndex:i] UTF8String];
        resultArray[i] = strdup(utf8Path);
    }

    return resultArray;
}

void freeStringArray(char** array, int count) {
    if (array == NULL) return;

    for (int i = 0; i < count; i++) {
        if (array[i] != NULL) {
            free(array[i]);
        }
    }
    free(array);
}

void freeString(char* str) {
    if (str != NULL) {
        free(str);
    }
}
*/
import "C"
import (
	"context"
	"fmt"
	"strings"
	"time"
	"unsafe"
)

// macOS Cocoa 实现
type darwinCocoaFinder struct{}

// 创建 macOS Cocoa 实现的工厂函数
func newFileFinder() FileFinder {
	return &darwinCocoaFinder{}
}

// 从剪贴板获取文件路径
func (f *darwinCocoaFinder) GetFiles(timeout time.Duration) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resultChan := make(chan []string, 1)
	errChan := make(chan error, 1)

	go func() {
		// 首先尝试直接获取文件路径
		var count C.int
		cPaths := C.getClipboardFilePaths(&count)

		// 如果找到文件路径，直接返回
		if count > 0 && cPaths != nil {
			defer C.freeStringArray(cPaths, count)

			// 将 C 字符串数组转换为 Go 字符串切片
			results := make([]string, int(count))
			for i := 0; i < int(count); i++ {
				cString := *(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cPaths)) + uintptr(i)*unsafe.Sizeof((*C.char)(nil))))
				results[i] = C.GoString(cString)
			}

			resultChan <- results
			return
		}

		// 如果没有直接的文件路径，尝试获取剪贴板文本
		cText := C.getClipboardText()
		if cText == nil {
			// 剪贴板为空，返回空结果
			resultChan <- []string{}
			return
		}

		// 确保释放内存
		defer C.freeString(cText)

		// 将剪贴板文本转换为文件名列表
		clipboardText := C.GoString(cText)
		fileNames := parseFileNames(clipboardText)

		if len(fileNames) == 0 {
			// 没有有效的文件名，返回空结果
			resultChan <- []string{}
			return
		}

		// 将 Go 字符串切片转换为 C 字符串数组
		cFileNames := make([]*C.char, len(fileNames))
		for i, name := range fileNames {
			cFileNames[i] = C.CString(name)
		}

		// 确保释放内存
		defer func() {
			for _, cName := range cFileNames {
				C.free(unsafe.Pointer(cName))
			}
		}()

		// 创建 C 字符串数组指针
		cFileNamesPtr := (**C.char)(unsafe.Pointer(&cFileNames[0]))

		// 根据文件名查找文件路径
		var resultCount C.int
		cResults := C.findFilesByNames(cFileNamesPtr, C.int(len(fileNames)), &resultCount)

		// 确保释放内存
		if cResults != nil {
			defer C.freeStringArray(cResults, resultCount)
		}

		// 将 C 字符串数组转换为 Go 字符串切片
		var results []string
		if resultCount > 0 && cResults != nil {
			results = make([]string, int(resultCount))
			for i := 0; i < int(resultCount); i++ {
				cString := *(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cResults)) + uintptr(i)*unsafe.Sizeof((*C.char)(nil))))
				results[i] = C.GoString(cString)
			}
		}

		resultChan <- results
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

// 解析剪贴板文本为文件名列表
func parseFileNames(text string) []string {
	if text == "" {
		return []string{}
	}

	// 分割文本为行
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")

	// 过滤空行
	var fileNames []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			fileNames = append(fileNames, line)
		}
	}

	return fileNames
}
