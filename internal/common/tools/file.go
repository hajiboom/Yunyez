// Package tools 文件操作通用工具函数
package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// 文件相关操作

//--------- 文件路径相关------------

// GetRootDir 获取项目根目录
// 从当前路径向上找 go.mod 所在目录
func GetRootDir() string {
	_, b, _, _ := runtime.Caller(0)
	// 向上找 go.mod 所在目录
	dir := filepath.Dir(b)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("project root not found (no go.mod)")
		}
		dir = parent
	}
}

// GetRuntimeDir 获取运行时目录
// 返回值:
//   - string: 运行时目录路径
//   - error: 处理过程中遇到的错误，若成功则为 nil
func GetRuntimeDir() (string, error) {
    execPath, err := os.Executable()
    if err != nil {
        return "", fmt.Errorf("failed to get executable path: %w", err)
    }
    // 二进制所在目录
    binDir := filepath.Dir(execPath)
    // configs 应该在 binDir/configs/
    return binDir, nil
}


// WriteFile 写入文件
// 参数：
//   - path: 文件路径
//   - data: 文件数据 字节切片
// 返回值:
//   - bool: 是否成功
//   - error: 处理过程中遇到的错误，若成功则为 nil
func WriteFile(path string, data []byte) (bool, error) {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return false, err
	}
	// 写入文件
	return true, os.WriteFile(path, data, 0644)
}
