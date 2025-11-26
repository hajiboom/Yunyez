package tools

import (
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

