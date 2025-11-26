package tools

import (
	"fmt"
	"testing"
)

func TestGetRootDir(t *testing.T) {
	rootDir := GetRootDir()
	fmt.Printf("项目根目录: %s\n", rootDir)
}
